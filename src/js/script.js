// Core Three.js objects
let scene, camera, renderer, controls, sphere;
let animationFrameId;

// Settings
let settings = {
  wireframe: document.getElementById('wireframe-toggle').checked,
  displacementScale: parseFloat(document.getElementById('displacement-slider').value),
  resolution: parseInt(document.getElementById('resolution-slider').value),
  latResolution: 256,
  lonResolution: 512,
  autoRotate: document.getElementById('autorotate-toggle').checked,
  color: document.getElementById('color-picker').value,
  planet: document.getElementById('planet-selection').value,
};

const geometryCache = new Map(); // Map<resolution, {geometry, originalPositions}>
const materialCache = new Map(); // Map<materialKey, material>

// Data management
let backendData = null; // Store latest backend response
let isApplyingDisplacement = false; // Prevent concurrent updates

// Performance monitoring
let frameCount = 0;
let lastFpsUpdate = performance.now();
let lastFrameTime = performance.now();

// INITIALIZATION
// ============================================================================

function init() {
  const container = document.getElementById('canvas-container');

  // Scene setup
  scene = new THREE.Scene();
  scene.background = new THREE.Color(0x0a0e1a);

  // Camera
  camera = new THREE.PerspectiveCamera(
    45,
    window.innerWidth / window.innerHeight,
    0.1,
    1000
  );
  camera.position.z = 5;

  // Renderer with optimizations
  renderer = new THREE.WebGLRenderer({
    antialias: true,
    powerPreference: "high-performance"
  });
  renderer.setSize(window.innerWidth, window.innerHeight);
  renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2));
  renderer.shadowMap.enabled = true;
  renderer.shadowMap.type = THREE.PCFSoftShadowMap;
  container.appendChild(renderer.domElement);

  // Controls
  controls = new THREE.OrbitControls(camera, renderer.domElement);
  controls.enableDamping = true;
  controls.dampingFactor = 0.05;
  controls.minDistance = 1.25;
  controls.maxDistance = 100.0;
  controls.autoRotate = false;
  controls.autoRotateSpeed = 0.5;
  controls.enablePan = true;
  controls.panSpeed = 0.8;
  controls.rotateSpeed = 0.8;
  controls.zoomSpeed = 1.0;

  setupLighting();
  createSphere(settings.latResolution, settings.lonResolution);
  setupEventListeners();
  animate();

  wireframeToggle(settings.wireframe);
  displacementChange(settings.displacementScale);
  resolutionChange(settings.resolution);
  autoRotateToggle(settings.autoRotate);
  colorChange(settings.color);
  planetChange(settings.planet);
}

// ============================================================================
// LIGHTING
// ============================================================================

function setupLighting() {
  const ambientLight = new THREE.AmbientLight(0xffffff, 0.4);
  scene.add(ambientLight);

  const sunLight = new THREE.DirectionalLight(0xffffff, 1.4);
  sunLight.position.set(5, 3, 5);
  sunLight.castShadow = true;
  sunLight.shadow.mapSize.width = 2048;
  sunLight.shadow.mapSize.height = 2048;
  scene.add(sunLight);

  const rimLight = new THREE.DirectionalLight(0x4a90e2, 0.6);
  rimLight.position.set(-5, 0, -5);
  scene.add(rimLight);
}

// ============================================================================
// GEOMETRY & MATERIAL CACHING SYSTEM
// ============================================================================

function getOrCreateBaseGeometry(resolution) {
  if (geometryCache.has(resolution)) {
    return geometryCache.get(resolution);
  }

  const geometry = new THREE.SphereGeometry(1, resolution, resolution);

  const cacheEntry = {
    geometry: geometry,
    vertexCount: geometry.attributes.position.count
  };

  geometryCache.set(resolution, cacheEntry);
  return cacheEntry;
}

function getOrCreateMaterial(wireframe, color) {
  const key = `wireframe_${wireframe};color_${color}`;

  if (materialCache.has(key)) {
    return materialCache.get(key);
  }

  const material = new THREE.MeshStandardMaterial({
    color: color,
    wireframe: wireframe,
    metalness: 0.3,
    roughness: 0.7,
    side: THREE.FrontSide
  });

  // Store uniforms here so we can update them globally without recompiling
  material.userData.uniforms = {
    uDisplacementScale: { value: settings.displacementScale },
    uDisplacementMap: { value: null },
    uDataRange: { value: new THREE.Vector2(-1, 1) } // For GPU normalization
  };

  material.onBeforeCompile = (shader) => {
    shader.uniforms.uDisplacementScale = material.userData.uniforms.uDisplacementScale;
    shader.uniforms.uDisplacementMap = material.userData.uniforms.uDisplacementMap;
    shader.uniforms.uDataRange = material.userData.uniforms.uDataRange;

    shader.vertexShader = `
      uniform float uDisplacementScale;
      uniform sampler2D uDisplacementMap;
      uniform vec2 uDataRange;
      ${shader.vertexShader}
    `;

    // Replace attribute mapping with UV texture lookup and on-the-fly normalization
    shader.vertexShader = shader.vertexShader.replace(
      '#include <begin_vertex>',
      `
      vec3 transformed = vec3(position);
      
      // Read raw displacement from the DataTexture
      float rawDisp = texture2D(uDisplacementMap, uv).r;
      
      // Normalize on the GPU: (val - mean) / halfRange
      float mean = (uDataRange.y + uDataRange.x) / 2.0;
      float halfRange = (uDataRange.y - uDataRange.x) / 2.0;
      float normalizedDisp = (halfRange == 0.0) ? 0.0 : (rawDisp - mean) / halfRange;
      
      transformed += normal * (normalizedDisp * uDisplacementScale);
      `
    );

    shader.fragmentShader = shader.fragmentShader.replace(
      '#include <normal_fragment_begin>',
      `
      vec3 fdx = dFdx(vViewPosition);
      vec3 fdy = dFdy(vViewPosition);
      vec3 normal = normalize(cross(fdx, fdy));
      vec3 geometryNormal = normal;
      `
    );
  };

  materialCache.set(key, material);
  return material;
}

// ============================================================================
// SPHERE CREATION
// Geometry only changes here and when the resolution slider is manually moved.
// ============================================================================

function createSphere(latResolution, lonResolution) {
  if (sphere) {
    scene.remove(sphere);
    sphere.traverse(child => {
      if (child.geometry) child.geometry.dispose();
    });
    sphere = null;
  }

  const material = getOrCreateMaterial(settings.wireframe, settings.color);
  sphere = new THREE.Group();

  const MAX_VERTICES_PER_SEGMENT = 1000000; // Lowered slightly for better compatibility
  const heightVertices = latResolution + 1;
  const maxWidthVertices = Math.floor(MAX_VERTICES_PER_SEGMENT / heightVertices) - 1;
  const segments = Math.max(1, Math.ceil(lonResolution / maxWidthVertices));

  for (let i = 0; i < segments; i++) {
    const phi_start = (Math.PI * 2 / segments) * i;
    const phi_length = Math.PI * 2 / segments;

    const geometry = new THREE.SphereGeometry(
      1,
      Math.ceil(lonResolution / segments),
      latResolution,
      phi_start,
      phi_length
    );

    const uvAttribute = geometry.attributes.uv;
    for (let j = 0; j < uvAttribute.count; j++) {
      let u = uvAttribute.getX(j);
      
      let globalU = (i / segments) + (u / segments);
      uvAttribute.setX(j, globalU);
    }

    const mesh = new THREE.Mesh(geometry, material);
    mesh.castShadow = true;
    mesh.receiveShadow = true;
    sphere.add(mesh);
  }

  scene.add(sphere);
  settings.latResolution = latResolution;
  settings.lonResolution = lonResolution;

  let totalVertices = 0;
  sphere.children.forEach(mesh => {
    totalVertices += mesh.geometry.attributes.position.count;
  });
  updateStats(totalVertices);
}
// ============================================================================
// DISPLACEMENT APPLICATION (GPU OPTIMIZED)
// Only called explicitly after a successful backend fetch
// ============================================================================

function applyBackendDisplacement(data) {
  if (!sphere || !data || !data.displacements || isApplyingDisplacement) return;

  isApplyingDisplacement = true;

  const width = data.lonResolution + 1;
  const height = data.latResolution + 1;
  const displacements = data.displacements;

  // 1. Find min/max for GPU normalization
  let min = Infinity;
  let max = -Infinity;
  for (let i = 0; i < displacements.length; i++) {
    const v = displacements[i];
    if (v < min) min = v;
    if (v > max) max = v;
  }

  // 2. Determine texture type based on array type
  const isFloat16 = typeof Float16Array !== 'undefined' && displacements instanceof Float16Array;
  const type = isFloat16 ? THREE.HalfFloatType : THREE.FloatType;

  const textureData = isFloat16 
    ? new Uint16Array(displacements.buffer, displacements.byteOffset, displacements.length)
    : displacements;

  // 3. Create the DataTexture directly from the raw array
  const texture = new THREE.DataTexture(
    textureData,
    width,
    height,
    THREE.RedFormat,
    type
  );
  
  texture.needsUpdate = true;
  texture.wrapS = THREE.RepeatWrapping; // Longitude wraps around
  texture.wrapT = THREE.ClampToEdgeWrapping; // Latitude clamps at poles
  texture.magFilter = THREE.LinearFilter;
  texture.minFilter = THREE.LinearFilter;
  texture.generateMipmaps = false;

  // 4. Pass the texture and min/max limits to the materials globally
  sphere.children.forEach(mesh => {
    if (mesh.material.userData.uniforms) {
      mesh.material.userData.uniforms.uDisplacementMap.value = texture;
      mesh.material.userData.uniforms.uDataRange.value.set(min, max);
    }
  });

  isApplyingDisplacement = false;
}

/**
 * Reset sphere vertices to the original smooth sphere positions.
 */
function resetToSmoothSphere() {
  if (!sphere) return;
  sphere.children.forEach(mesh => {
    if (mesh.material.userData.uniforms) {
      mesh.material.userData.uniforms.uDisplacementMap.value = null;
    }
  });
}

// ============================================================================
// ANIMATION LOOP
// ============================================================================

function animate() {
  animationFrameId = requestAnimationFrame(animate);

  const currentTime = performance.now();

  frameCount++;
  if (currentTime - lastFpsUpdate > 1000) {
    const fps = Math.round((frameCount * 1000) / (currentTime - lastFpsUpdate));
    document.getElementById('fps-stat').textContent = fps;
    frameCount = 0;
    lastFpsUpdate = currentTime;
  }

  controls.update();

  const distance = camera.position.length();
  document.getElementById('zoom-stat').textContent = distance.toFixed(2);

  renderer.render(scene, camera);
}

// ============================================================================
// BACKEND COMMUNICATION
// ============================================================================

async function fetchTopographyData() {
  const btn = document.getElementById('fetch-btn');
  const url = "/topography";

  btn.disabled = true;
  btn.textContent = 'Loading…';

  fetch: try {
    if (backendData && settings.resolution == backendData.resolution && backendData.planet == settings.planet) {
      await new Promise(resolve => setTimeout(resolve, 250));
      break fetch;
    }

    const response = await fetch(`${url}?src=${encodeURIComponent(settings.planet)}&res=${encodeURIComponent(settings.resolution)}`, {
      method: 'GET'
    });

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`);
    }

    const buffer = await response.arrayBuffer();
    const view = new DataView(buffer);

    const dataType = view.getUint32(0, true);
    const vertexCount = view.getUint32(4, true);
    const latPoints = view.getUint32(8, true);
    const lonPoints = view.getUint32(12, true);

    let displacements;
    if (dataType == 0) {
      // float16
      displacements = new Float16Array(buffer, 16, vertexCount);
    } else if (dataType == 1) {
      // float32
      displacements = new Float32Array(buffer, 16, vertexCount);
    }

    backendData = {
      displacements: displacements,
      resolution: settings.resolution,
      planet: settings.planet,
      latResolution: latPoints - 1,
      lonResolution: lonPoints - 1,
      metadata: {
        vertex_count: vertexCount,
        lat_points: latPoints,
        lon_points: lonPoints
      },
    };

    createSphere(backendData.latResolution, backendData.lonResolution);
    applyBackendDisplacement(backendData);
  } catch (error) {
    alert(`Backend connection failed:\n${error.message}`);
  } finally {
    btn.disabled = false;
    btn.textContent = 'Fetch Topography Data';
  }
}

// ============================================================================
// EVENT HANDLERS
// ============================================================================

function setupEventListeners() {
  window.addEventListener('resize', onWindowResize);
  document.getElementById('fetch-btn').addEventListener('click', fetchTopographyData);

  document.getElementById('planet-selection').addEventListener('change', onPlanetChange);
  document.getElementById('displacement-slider').addEventListener('input', onDisplacementChange);
  document.getElementById('resolution-slider').addEventListener('input', onResolutionChange);
  document.getElementById('wireframe-toggle').addEventListener('change', onWireframeToggle);
  document.getElementById('autorotate-toggle').addEventListener('change', onAutoRotateToggle);
  document.getElementById('color-picker').addEventListener('input', onColorChange);
}

function onWindowResize() {
  camera.aspect = window.innerWidth / window.innerHeight;
  camera.updateProjectionMatrix();
  renderer.setSize(window.innerWidth, window.innerHeight);
}

function onPlanetChange(e) { planetChange(e.target.value); }
function onDisplacementChange(e) { displacementChange(parseFloat(e.target.value)); }
function onResolutionChange(e) { resolutionChange(parseInt(e.target.value)); }
function onWireframeToggle(e) { wireframeToggle(e.target.checked); }
function onAutoRotateToggle(e) { autoRotateToggle(e.target.checked); }
function onColorChange(e) { colorChange(e.target.value); }

function planetChange(v) { 
  settings.planet = v 
}

function displacementChange(v) { 
  settings.displacementScale = v;
  document.getElementById('displacement-value').textContent = settings.displacementScale.toFixed(2);

  if (sphere) {
    sphere.children.forEach(mesh => {
      if (mesh.material.userData.uniforms) {
        mesh.material.userData.uniforms.uDisplacementScale.value = settings.displacementScale;
      }
    });
  }
}

function resolutionChange(v) {
  document.getElementById('resolution-value').textContent = v;
  settings.resolution = v;
}

function wireframeToggle(v) {
  settings.wireframe = v;
  if (sphere) {
    const newMat = getOrCreateMaterial(settings.wireframe, settings.color);
    sphere.children.forEach(mesh => {
      mesh.material = newMat;
      
      // Preserve displacement state when swapping materials
      if (backendData && !isApplyingDisplacement) {
        applyBackendDisplacement(backendData);
      }
    });
  }
}

function autoRotateToggle(v) {
  settings.autoRotate = v;
  controls.autoRotate = v;
}

function colorChange(v) {
  settings.color = parseInt(v.replace('#', ''), 16);

  if (sphere) {
    const newMat = getOrCreateMaterial(settings.wireframe, settings.color);
    sphere.children.forEach(mesh => {
      mesh.material = newMat;
      
      // Preserve displacement state when swapping materials
      if (backendData && !isApplyingDisplacement) {
        applyBackendDisplacement(backendData);
      }
    });
  }
}

// ============================================================================
// UI UPDATES
// ============================================================================

function updateStats(vertexCount) {
  document.getElementById('verts-stat').textContent = vertexCount.toLocaleString();
}

// ============================================================================
// CLEANUP
// ============================================================================

function cleanup() {
  if (animationFrameId) cancelAnimationFrame(animationFrameId);

  for (const [, entry] of geometryCache) entry.geometry.dispose();
  geometryCache.clear();

  for (const [, material] of materialCache) material.dispose();
  materialCache.clear();

  if (renderer) renderer.dispose();
  if (controls) controls.dispose();
}

window.addEventListener('beforeunload', cleanup);

// ============================================================================
// BOOT
// ============================================================================

if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', init);
} else {
  init();
}

// ============================================================================
// DEBUG UTILITIES
// ============================================================================

window.earthViewer = {
  getStats: () => ({
    resolution: settings.resolution,
    cacheSize: geometryCache.size,
    materialCacheSize: materialCache.size,
    vertexCount: sphere?.geometry.attributes.position.count ?? 0,
    hasBackendData: !!backendData,
    displacementScale: settings.displacementScale,
    dataStats: backendData?.metadata ?? null
  }),

  clearCache: () => {
    for (const [, entry] of geometryCache) entry.geometry.dispose();
    geometryCache.clear();
  },

  resetSphere: () => {
    backendData = null;
    resetToSmoothSphere();
  },
};