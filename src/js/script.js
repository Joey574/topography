// Core Three.js objects
let scene, camera, renderer, controls, sphere;
let animationFrameId;

// Settings
let settings = {
  wireframe: false,
  displacementScale: 0.10,
  resolution: 512,
  latResolution: 256,
  lonResolution: 512,
  autoRotate: false,
  color: 0x2a5784,
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
// GEOMETRY CACHING SYSTEM
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
    uDisplacementScale: { value: settings.displacementScale }
  };

  material.onBeforeCompile = (shader) => {
  shader.uniforms.uDisplacementScale = material.userData.uniforms.uDisplacementScale;

  shader.vertexShader = `
    attribute float displacement;
    uniform float uDisplacementScale;
    ${shader.vertexShader}
  `;

  shader.vertexShader = shader.vertexShader.replace(
    '#include <begin_vertex>',
    `
    vec3 transformed = vec3(position);
    transformed += normal * (displacement * uDisplacementScale);
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
// Displacement is ONLY applied via the fetch button.
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
  sphere.castShadow = true;
  sphere.receiveShadow = true;

  const MAX_VERTICES_PER_SEGMENT = 5000000;
  const heightVertices = latResolution+1;
  const maxWidthVertices = Math.floor(MAX_VERTICES_PER_SEGMENT / heightVertices) - 1;
  const segments = Math.max(1, Math.ceil(lonResolution / maxWidthVertices));

  for (let i = 0; i < segments; i++) {
    const phi_start = (Math.PI * 2 / segments) * i;
    const phi_length = Math.PI * 2 / segments;

    const geometry = new THREE.SphereGeometry(
      1,
      lonResolution / segments,
      latResolution,
      phi_start,
      phi_length
    )

    const mesh = new THREE.Mesh(geometry, material);
    mesh.userData.segmentIndex = i;
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
// DATA NORMALIZATION
// ============================================================================

function normalizeDisplacements(displacements) {
  if (!displacements || displacements.length === 0) return displacements;

  // Find min and max
  let min = Infinity;
  let max = -Infinity;
  
  for (let i = 0; i < displacements.length; i++) {
    const val = displacements[i];
    if (val < min) min = val;
    if (val > max) max = val;
  }

  const range = max - min;
  const normalized = new Float32Array(displacements.length);

  if (range === 0) {
    // Flat data - return zeros
    console.warn('Displacement data has no variation (flat surface)');
    return normalized;
  }

  // Normalize to [-1, 1] range centered at mean
  const mean = (max + min) / 2;
  const halfRange = range / 2;

  for (let i = 0; i < displacements.length; i++) {
    normalized[i] = (displacements[i] - mean) / halfRange;
  }

  return normalized;
}

// ============================================================================
// DISPLACEMENT APPLICATION
// Only called explicitly after a successful backend fetch
// ============================================================================

function applyBackendDisplacement(data) {
  if (!sphere || !data || !data.displacements || isApplyingDisplacement) return;

  isApplyingDisplacement = true;

  const latResolution = data.latResolution;
  const lonResolution = data.lonResolution;
  
  // Normalize the raw displacement data first
  const normalizedDisplacements = normalizeDisplacements(data.displacements)

  const latDivisions = latResolution + 1;
  const lonDivisions = lonResolution + 1;

  sphere.children.forEach((mesh, segmentIndex) => {
    const geometry = mesh.geometry;
    const positions = geometry.attributes.position;
    const displacementArray = new Float32Array(positions.count);

    for (let i = 0; i < positions.count; i++) {
      const x = positions.getX(i);
      const y = positions.getY(i);
      const z = positions.getZ(i);

      const phi = Math.atan2(x, z);
      const theta = Math.acos(Math.max(-1, Math.min(1, y)));

      const latIndex = Math.round((theta / Math.PI) * latResolution);
      const lonIndex = Math.round(((phi + Math.PI) / (Math.PI * 2)) * lonResolution) % lonDivisions;

      const backendIndex = latIndex * lonDivisions + lonIndex;
      displacementArray[i] = normalizedDisplacements[backendIndex] || 0;
    }

    geometry.setAttribute('displacement', new THREE.BufferAttribute(displacementArray, 1));
  });

  isApplyingDisplacement = false;
}

/**
 * Reset sphere vertices to the original smooth sphere positions.
 * Called when displacement scale is changed but no backend data exists,
 * or when manually requested.
 */
function resetToSmoothSphere() {
  if (!sphere) return;
  sphere.children.forEach(mesh => {
    mesh.geometry.deleteAttribute('displacement');
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
  const url = "/topography"

  btn.disabled = true;
  btn.textContent = 'Loading…';

  fetch: try {
    // ensure we only request when resolution has changed, sleep to give impression of work done
    if (backendData && backendData.resolution && settings.resolution == backendData.resolution) {
      await new Promise(resolve => setTimeout(resolve, 250));
      break fetch;
    }

    const response = await fetch(`${url}?res=${encodeURIComponent(settings.resolution)}`, {
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
    } else if (dataType == 2) {
      // float32
      displacements = new Float32Array(buffer, 16, vertexCount);
    }

    backendData = {
      displacements: displacements,
      resolution: settings.resolution,
      latResolution: latPoints - 1,
      lonResolution: lonPoints - 1,
      metadata: {
        vertex_count: vertexCount,
        lat_points: latPoints,
        lon_points: lonPoints
      },
    };

    // we recreate geometry only when a fetch has been performed
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

function onDisplacementChange(e) {
  settings.displacementScale = parseFloat(e.target.value);
  document.getElementById('displacement-value').textContent = settings.displacementScale.toFixed(2);

  if (sphere) {
    sphere.children.forEach(mesh => {
      if (mesh.material.userData.uniforms) {
        mesh.material.userData.uniforms.uDisplacementScale.value = settings.displacementScale;
      }
    });
  }
}

function onResolutionChange(e) {
  const newResolution = parseInt(e.target.value);
  document.getElementById('resolution-value').textContent = newResolution;

  // we don't recreate geometry here, so the user
  // doesn't experience slowdown until requesting
  settings.resolution = newResolution
}

function onWireframeToggle(e) {
  settings.wireframe = e.target.checked;
  if (sphere) {
    const newMat = getOrCreateMaterial(settings.wireframe, settings.color);
    sphere.children.forEach(mesh => {
      mesh.material = newMat;
    });
  }
}

function onAutoRotateToggle(e) {
  settings.autoRotate = e.target.checked;
  controls.autoRotate = settings.autoRotate;
}

function onColorChange(e) {
  // Convert hex string (#rrggbb) to integer
  const hexStr = e.target.value;
  settings.color = parseInt(hexStr.replace('#', ''), 16);

  if (sphere) {
    const newMat = getOrCreateMaterial(settings.wireframe, settings.color);
    sphere.children.forEach(mesh => {
      mesh.material = newMat;
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
    dataStats: backendData?.stats ?? null
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
