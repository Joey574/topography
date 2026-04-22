// Core Three.js objects
let scene, camera, renderer, controls, sphereGroup;
let animationFrameId;

let settings = {
  displacementScale: 0.10,
  resolution: 128,
  color: 0x2a5784,
  wireframe: false,
  inverted: true // Added toggle for inversion
};

let backendData = null;

function init() {
  const container = document.getElementById('canvas-container');
  scene = new THREE.Scene();
  scene.background = new THREE.Color(0x0a0e1a);

  camera = new THREE.PerspectiveCamera(45, window.innerWidth / window.innerHeight, 0.1, 1000);
  camera.position.z = 5;

  renderer = new THREE.WebGLRenderer({ antialias: true, powerPreference: "high-performance" });
  renderer.setSize(window.innerWidth, window.innerHeight);
  renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2));
  container.appendChild(renderer.domElement);

  controls = new THREE.OrbitControls(camera, renderer.domElement);
  controls.enableDamping = true;

  setupLighting();
  createSegmentedSphere(128, 128); 
  setupEventListeners();
  animate();
}

function setupLighting() {
  scene.add(new THREE.AmbientLight(0xffffff, 0.4));
  const sunLight = new THREE.DirectionalLight(0xffffff, 1.2);
  sunLight.position.set(5, 3, 5);
  scene.add(sunLight);
}

// ============================================================================
// SEGMENTED GEOMETRY
// ============================================================================

function createSegmentedSphere(lonRes, latRes) {
  if (sphereGroup) {
    scene.remove(sphereGroup);
    sphereGroup.traverse(child => {
      if (child.geometry) child.geometry.dispose();
    });
  }

  sphereGroup = new THREE.Group();
  
  // Calculate segments based on Index Count limit
  // A quad (2 triangles) uses 6 indices. 
  const totalIndices = lonRes * latRes * 6;
  const MAX_INDICES = 25000000; // Staying safely under your 30M limit
  const numSegments = Math.ceil(totalIndices / MAX_INDICES);
  
  const lonPerSegment = lonRes / numSegments;

  const material = new THREE.MeshStandardMaterial({
    color: settings.color,
    wireframe: settings.wireframe,
    metalness: 0.1,
    roughness: 0.8,
  });

  // Shader injection for high-performance normals and inversion
  material.onBeforeCompile = (shader) => {
    shader.uniforms.uDisplacementScale = { value: settings.displacementScale };
    material.userData.shader = shader;

    shader.vertexShader = shader.vertexShader.replace(
      '#include <common>',
      `#include <common>
       uniform float uDisplacementScale;`
    );

    // Standard displacementMap uses 'vUv'. If your data is inverted vertically,
    // we can flip it here or in the data normalization.
    shader.fragmentShader = shader.fragmentShader.replace(
      '#include <normal_fragment_begin>',
      `
      vec3 fdx = dFdx( vViewPosition );
      vec3 fdy = dFdy( vViewPosition );
      vec3 normal = normalize( cross( fdx, fdy ) );
      vec3 geometryNormal = normal;
      `
    );
  };

  for (let i = 0; i < numSegments; i++) {
    const phiStart = (i / numSegments) * Math.PI * 2;
    const phiLength = (1 / numSegments) * Math.PI * 2;

    // Use partial SphereGeometry. UVs are automatically calculated 0-1 for the whole sphere.
    const geometry = new THREE.SphereGeometry(1, Math.ceil(lonPerSegment), latRes, phiStart, phiLength);
    const mesh = new THREE.Mesh(geometry, material);
    sphereGroup.add(mesh);
  }

  scene.add(sphereGroup);
  document.getElementById('verts-stat').textContent = (lonRes * latRes).toLocaleString();
}

// ============================================================================
// DATA & TEXTURE
// ============================================================================

function updateDisplacementTexture(data) {
  if (!sphereGroup || !data) return;

  const width = data.lonPoints;
  const height = data.latPoints;
  
  // FIX: Inversion. If displacements were "inverting" the terrain (valleys as peaks),
  // we flip the normalization math.
  const normalized = normalizeDisplacements(data.displacements, settings.inverted);

  const texture = new THREE.DataTexture(
    normalized,
    width,
    height,
    THREE.RedFormat,
    THREE.FloatType
  );
  
  texture.minFilter = THREE.LinearFilter;
  texture.magFilter = THREE.LinearFilter;
  texture.wrapS = THREE.RepeatWrapping; 
  texture.needsUpdate = true;

  // Apply to the material shared by all segments
  const sharedMaterial = sphereGroup.children[0].material;
  sharedMaterial.displacementMap = texture;
  sharedMaterial.displacementScale = settings.displacementScale;
  sharedMaterial.needsUpdate = true;
}

function normalizeDisplacements(displacements, inverted) {
  let min = Infinity, max = -Infinity;
  for (let i = 0; i < displacements.length; i++) {
    if (displacements[i] < min) min = displacements[i];
    if (displacements[i] > max) max = displacements[i];
  }
  const range = max - min;
  const normalized = new Float32Array(displacements.length);

  for (let i = 0; i < displacements.length; i++) {
    let val = (displacements[i] - min) / (range || 1);
    // If inverted, 1 becomes 0 and 0 becomes 1
    normalized[i] = inverted ? 1.0 - val : val;
  }
  return normalized;
}

// ============================================================================
// FETCH & ANIMATE
// ============================================================================

async function fetchTopographyData() {
  const btn = document.getElementById('fetch-btn');
  btn.disabled = true;
  btn.textContent = 'Processing...';

  try {
    const response = await fetch("/topography", {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ resolution: settings.resolution })
    });

    const buffer = await response.arrayBuffer();
    const view = new DataView(buffer);
    const dataType = view.getUint32(0, true);
    const vertexCount = view.getUint32(4, true);
    const latPoints = view.getUint32(8, true);
    const lonPoints = view.getUint32(12, true);

    const displacements = (dataType === 2) 
      ? new Float32Array(buffer, 16, vertexCount)
      : new Float32Array(new Uint16Array(buffer, 16, vertexCount)); // basic float16 fallback

    backendData = { displacements, latPoints, lonPoints };

    // Re-segment based on the new resolution
    createSegmentedSphere(lonPoints - 1, latPoints - 1);
    updateDisplacementTexture(backendData);

  } catch (error) {
    console.error("Fetch error:", error);
  } finally {
    btn.disabled = false;
    btn.textContent = 'Fetch Topography Data';
  }
}

function animate() {
  animationFrameId = requestAnimationFrame(animate);
  controls.update();
  renderer.render(scene, camera);
}

function setupEventListeners() {
  window.addEventListener('resize', () => {
    camera.aspect = window.innerWidth / window.innerHeight;
    camera.updateProjectionMatrix();
    renderer.setSize(window.innerWidth, window.innerHeight);
  });
  
  document.getElementById('fetch-btn').addEventListener('click', fetchTopographyData);
  
  document.getElementById('displacement-slider').addEventListener('input', (e) => {
    settings.displacementScale = parseFloat(e.target.value);
    sphereGroup.children[0].material.displacementScale = settings.displacementScale;
    if (sphereGroup.children[0].material.userData.shader) {
      sphereGroup.children[0].material.userData.shader.uniforms.uDisplacementScale.value = settings.displacementScale;
    }
  });

  document.getElementById('resolution-slider').addEventListener('input', (e) => {
    settings.resolution = parseInt(e.target.value);
  });
}

init();