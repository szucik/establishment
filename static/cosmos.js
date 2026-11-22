// Animated deep-space background: nebulae, twinkling stars and a slowly rotating spiral galaxy.
(function () {
  const STAR_COUNT = 450;
  const GALAXY_PARTICLES = 7000;
  const GALAXY_ARMS = 2;
  const ARM_WINDING_RADIANS = 4.2;
  const ARM_SCATTER = 0.55;
  const CORE_FRACTION = 0.18;
  const ROTATION_PER_FRAME = 0.0003;
  const GALAXY_TILT = 0.5;

  const prefersReducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;

  const canvas = document.createElement('canvas');
  canvas.setAttribute('aria-hidden', 'true');
  Object.assign(canvas.style, { position: 'fixed', inset: '0', width: '100%', height: '100%', zIndex: '-1', pointerEvents: 'none' });
  document.documentElement.style.background = '#02030a';
  document.body.style.background = 'transparent';
  document.body.prepend(canvas);

  const context = canvas.getContext('2d');
  const backdrop = document.createElement('canvas');
  let stars = [];
  let galaxy = [];
  let rotation = 0;

  function gaussian() {
    return (Math.random() + Math.random() + Math.random() - 1.5) / 1.5;
  }

  function createStars(width, height) {
    return Array.from({ length: STAR_COUNT }, () => ({
      x: Math.random() * width,
      y: Math.random() * height,
      radius: Math.random() < 0.92 ? Math.random() * 0.9 + 0.2 : Math.random() * 1.6 + 0.8,
      phase: Math.random() * Math.PI * 2,
      speed: 0.01 + Math.random() * 0.03,
    }));
  }

  function createCoreParticle() {
    return {
      distance: Math.abs(gaussian()) * 0.16,
      angle: Math.random() * Math.PI * 2,
      size: Math.random() * 1.4 + 0.5,
      color: `rgba(255, ${215 + Math.floor(Math.random() * 40)}, 190, ${0.35 + Math.random() * 0.5})`,
    };
  }

  function createArmParticle() {
    const distance = 0.08 + Math.pow(Math.random(), 0.9) * 0.92;
    const arm = Math.floor(Math.random() * GALAXY_ARMS);
    // Scatter grows outwards so arms are tight near the core and diffuse at the rim.
    const scatter = gaussian() * ARM_SCATTER * (0.25 + distance * 0.6);
    const roll = Math.random();
    return {
      distance: distance + gaussian() * 0.03,
      angle: arm * (2 * Math.PI / GALAXY_ARMS) + distance * ARM_WINDING_RADIANS + scatter,
      size: Math.random() * 1.3 + 0.3,
      color: roll < 0.12 ? `rgba(255, 140, 220, ${0.35 + Math.random() * 0.4})`
        : roll < 0.3 ? `rgba(120, 255, 214, ${0.25 + Math.random() * 0.35})`
        : `rgba(170, 205, 255, ${0.3 + Math.random() * 0.5})`,
    };
  }

  function createGalaxy() {
    return Array.from({ length: GALAXY_PARTICLES },
      () => (Math.random() < CORE_FRACTION ? createCoreParticle() : createArmParticle()));
  }

  function paintNebula(ctx, x, y, radius, color) {
    const gradient = ctx.createRadialGradient(x, y, 0, x, y, radius);
    gradient.addColorStop(0, color);
    gradient.addColorStop(1, 'rgba(0, 0, 0, 0)');
    ctx.fillStyle = gradient;
    ctx.fillRect(0, 0, ctx.canvas.width, ctx.canvas.height);
  }

  function paintBackdrop(width, height) {
    backdrop.width = width;
    backdrop.height = height;
    const ctx = backdrop.getContext('2d');
    const space = ctx.createRadialGradient(width / 2, height / 2, 0, width / 2, height / 2, Math.max(width, height) * 0.75);
    space.addColorStop(0, '#0a0f2e');
    space.addColorStop(1, '#02030a');
    ctx.fillStyle = space;
    ctx.fillRect(0, 0, width, height);
    const size = Math.max(width, height);
    paintNebula(ctx, width * 0.12, height * 0.18, size * 0.42, 'rgba(109, 40, 217, 0.35)');
    paintNebula(ctx, width * 0.88, height * 0.82, size * 0.4, 'rgba(37, 99, 235, 0.3)');
    paintNebula(ctx, width * 0.78, height * 0.12, size * 0.28, 'rgba(219, 39, 119, 0.2)');
    paintNebula(ctx, width * 0.2, height * 0.9, size * 0.3, 'rgba(8, 145, 178, 0.22)');
  }

  function drawGalaxy(width, height) {
    const centerX = width * 0.6;
    const centerY = height * 0.52;
    const radius = Math.min(width, height) * 0.6;

    context.save();
    context.translate(centerX, centerY);
    context.scale(1, GALAXY_TILT);
    const halo = context.createRadialGradient(0, 0, 0, 0, 0, radius * 0.7);
    halo.addColorStop(0, 'rgba(255, 236, 205, 0.55)');
    halo.addColorStop(0.12, 'rgba(255, 200, 170, 0.22)');
    halo.addColorStop(0.45, 'rgba(140, 120, 255, 0.08)');
    halo.addColorStop(1, 'rgba(0, 0, 0, 0)');
    context.fillStyle = halo;
    context.fillRect(-radius, -radius, radius * 2, radius * 2);
    context.restore();

    context.globalCompositeOperation = 'lighter';
    for (const particle of galaxy) {
      const angle = particle.angle + rotation * (1.4 - particle.distance);
      const x = centerX + Math.cos(angle) * particle.distance * radius;
      const y = centerY + Math.sin(angle) * particle.distance * radius * GALAXY_TILT;
      context.fillStyle = particle.color;
      context.fillRect(x, y, particle.size, particle.size);
    }
    context.globalCompositeOperation = 'source-over';
  }

  function drawStars() {
    for (const star of stars) {
      const brightness = 0.55 + 0.45 * Math.sin(star.phase);
      context.globalAlpha = brightness;
      context.fillStyle = '#ffffff';
      context.beginPath();
      context.arc(star.x, star.y, star.radius, 0, Math.PI * 2);
      context.fill();
      if (!prefersReducedMotion) star.phase += star.speed;
    }
    context.globalAlpha = 1;
  }

  function drawFrame() {
    context.drawImage(backdrop, 0, 0);
    drawGalaxy(canvas.width, canvas.height);
    drawStars();
  }

  function animate() {
    rotation += ROTATION_PER_FRAME;
    drawFrame();
    requestAnimationFrame(animate);
  }

  function resize() {
    canvas.width = window.innerWidth;
    canvas.height = window.innerHeight;
    paintBackdrop(canvas.width, canvas.height);
    stars = createStars(canvas.width, canvas.height);
    if (prefersReducedMotion) drawFrame();
  }

  galaxy = createGalaxy();
  resize();
  window.addEventListener('resize', resize);
  if (prefersReducedMotion) drawFrame(); else requestAnimationFrame(animate);
})();
