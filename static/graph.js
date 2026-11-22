// "Nici powiązań" — the neon relationship graph shared by the home page and the panel.
// Returns a user-facing warning ('' when everything rendered).
// With { focusId } it draws one person's network: that person pinned in the centre, their connections around.
(function () {
  const GRAPH_HEIGHT = 600;
  const AVATAR_RADIUS = 24;
  const POLITICIAN = 'Polityk';
  const FOCUS_ORBIT_RADIUS = 190;
  const HALO_RADIUS = AVATAR_RADIUS * 1.9;

  // Matte planet colours [lit side, body, shadow side]; each person always gets the same one.
  const PLANET_PALETTES = [
    ['#6f8fb5', '#4a6a94', '#1c2b45'],
    ['#8a78b0', '#64548c', '#271f3f'],
    ['#ad7a95', '#875670', '#3a2030'],
    ['#b89a68', '#8f7447', '#3d2f1a'],
    ['#6fa39a', '#4b7d75', '#1b3431'],
    ['#7a80b0', '#555b8a', '#20223f'],
  ];
  const STAR_PALETTE = ['#e8d9b0', '#c9a865', '#6b4f22'];

  let stopPrevious = () => {};

  const escapeHtml = (text) => String(text ?? '').replace(/[&<>"']/g,
    (char) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' })[char]);

  function validEdgesOf(graph) {
    const nodeIds = new Set(graph.nodes.map((node) => node.id));
    return graph.edges
      .filter((edge) => nodeIds.has(edge.source_id) && nodeIds.has(edge.target_id) && edge.source_id !== edge.target_id)
      .map((edge) => ({ source: edge.source_id, target: edge.target_id, type: edge.type, details: edge.details }));
  }

  function focusedOn(graph, personId) {
    const edges = graph.edges.filter((edge) => edge.source_id === personId || edge.target_id === personId);
    const connectedIds = new Set([personId, ...edges.flatMap((edge) => [edge.source_id, edge.target_id])]);
    return { nodes: graph.nodes.filter((node) => connectedIds.has(node.id)), edges };
  }

  function paletteIndexOf(id) {
    let hash = 0;
    for (const char of id || '') hash = (hash * 31 + char.charCodeAt(0)) >>> 0;
    return hash % PLANET_PALETTES.length;
  }

  function planetPaletteOf(id) {
    return PLANET_PALETTES[paletteIndexOf(id)];
  }

  function initialsOf(name) {
    return (name || '?').split(/\s+/).filter(Boolean).slice(0, 2).map((word) => word[0].toUpperCase()).join('');
  }

  function tooltipHtmlOf(person) {
    return `
      <strong>${escapeHtml(person.name)}</strong><br>
      Zawód: ${escapeHtml(person.occupation)}<br>
      ${person.description ? 'Opis: ' + escapeHtml(person.description) + '<br>' : ''}
      ${person.twitter ? 'Twitter: ' + escapeHtml(person.twitter) : ''}`;
  }

  function ensureTooltip() {
    const existing = d3.select('body').select('.tooltip');
    return existing.empty() ? d3.select('body').append('div').attr('class', 'tooltip').style('opacity', 0) : existing;
  }

  function appendSphereGradient(defs, id, [highlight, body, shadow]) {
    const sphere = defs.append('radialGradient').attr('id', id).attr('cx', '40%').attr('cy', '38%').attr('r', '80%');
    sphere.append('stop').attr('offset', '0%').attr('stop-color', highlight);
    sphere.append('stop').attr('offset', '55%').attr('stop-color', body);
    sphere.append('stop').attr('offset', '100%').attr('stop-color', shadow);
  }

  function appendHaloGradient(defs, id, color) {
    const halo = defs.append('radialGradient').attr('id', id);
    halo.append('stop').attr('offset', '45%').attr('stop-color', color).attr('stop-opacity', 0.14);
    halo.append('stop').attr('offset', '100%').attr('stop-color', color).attr('stop-opacity', 0);
  }

  function appendDefinitions(svg) {
    const defs = svg.append('defs');
    defs.append('clipPath').attr('id', 'circle-clip')
      .append('circle').attr('cx', 0).attr('cy', 0).attr('r', AVATAR_RADIUS);

    const beam = defs.append('linearGradient').attr('id', 'aurora-gradient')
      .attr('x1', '0%').attr('y1', '0%').attr('x2', '100%').attr('y2', '100%');
    beam.append('stop').attr('offset', '0%').attr('stop-color', '#7f93b5');
    beam.append('stop').attr('offset', '50%').attr('stop-color', '#9a8fb5');
    beam.append('stop').attr('offset', '100%').attr('stop-color', '#7fa8a0');

    PLANET_PALETTES.forEach((palette, index) => {
      appendSphereGradient(defs, `planet-${index}`, palette);
      appendHaloGradient(defs, `halo-${index}`, palette[1]);
    });
    appendSphereGradient(defs, 'star', STAR_PALETTE);
    appendHaloGradient(defs, 'halo-star', STAR_PALETTE[1]);
  }

  function appendCelestialBody(node, focusNode) {
    const isStar = (d) => d === focusNode;
    node.append('circle').attr('class', 'halo')
      .attr('r', (d) => (isStar(d) ? HALO_RADIUS * 1.5 : HALO_RADIUS))
      .attr('fill', (d) => (isStar(d) ? 'url(#halo-star)' : `url(#halo-${paletteIndexOf(d.id)})`));
    node.append('circle').attr('class', 'planet')
      .attr('r', AVATAR_RADIUS)
      .attr('fill', (d) => (isStar(d) ? 'url(#star)' : `url(#planet-${paletteIndexOf(d.id)})`));
    node.filter((d) => d.image_url).append('image')
      .attr('href', (d) => d.image_url)
      .attr('x', -AVATAR_RADIUS).attr('y', -AVATAR_RADIUS)
      .attr('width', AVATAR_RADIUS * 2).attr('height', AVATAR_RADIUS * 2)
      .attr('preserveAspectRatio', 'xMidYMid slice')
      .attr('clip-path', 'url(#circle-clip)')
      .on('error', function () { this.remove(); });
    node.filter((d) => !d.image_url).append('text').attr('class', 'initials')
      .attr('text-anchor', 'middle').attr('dy', '.35em')
      .text((d) => initialsOf(d.name));
    node.append('circle').attr('class', (d) => (isStar(d) ? 'node-circle node-focus' : 'node-circle'))
      .attr('r', (d) => (isStar(d) ? AVATAR_RADIUS + 6 : AVATAR_RADIUS + 1)).attr('fill', 'none');
  }

  function dragBehaviour(simulation) {
    return d3.drag()
      .on('start', (event, d) => {
        if (!event.active) simulation.alphaTarget(0.3).restart();
        d.fx = d.x; d.fy = d.y;
      })
      .on('drag', (event, d) => { d.fx = event.x; d.fy = event.y; })
      .on('end', (event, d) => {
        if (!event.active) simulation.alphaTarget(0);
        d.fx = d.pinX ?? null; d.fy = d.pinY ?? null;
      });
  }

  function groupForces(simulation, width) {
    simulation
      .force('center', d3.forceCenter(width / 2, GRAPH_HEIGHT / 2).strength(0.2))
      .force('x', d3.forceX().x((d) => (d.occupation === POLITICIAN ? width / 3 : 2 * width / 3)).strength(0.1));
  }

  function orbitForces(simulation, width, focusNode) {
    focusNode.pinX = focusNode.fx = width / 2;
    focusNode.pinY = focusNode.fy = GRAPH_HEIGHT / 2;
    simulation.force('orbit', d3.forceRadial(FOCUS_ORBIT_RADIUS, width / 2, GRAPH_HEIGHT / 2)
      .strength((d) => (d === focusNode ? 0 : 0.8)));
  }

  function renderRelationsGraph(selector, fullGraph, { focusId } = {}) {
    stopPrevious();
    const container = d3.select(selector);
    container.selectAll('*').remove();
    const graph = focusId ? focusedOn(fullGraph, focusId) : fullGraph;
    if (!graph.nodes.length) return 'Brak osób do wyświetlenia w grafie';

    const nodes = graph.nodes.map((person) => ({ ...person }));
    const edges = validEdgesOf(graph);
    const width = container.node().getBoundingClientRect().width;
    const focusNode = nodes.find((node) => node.id === focusId);
    const applyLayout = (currentWidth) => (focusNode
      ? orbitForces(simulation, currentWidth, focusNode)
      : groupForces(simulation, currentWidth));

    const svg = container.append('svg').attr('width', '100%').attr('height', GRAPH_HEIGHT);
    appendDefinitions(svg);

    const simulation = d3.forceSimulation(nodes)
      .force('link', d3.forceLink(edges).id((d) => d.id).distance(120).strength(0.15))
      .force('charge', d3.forceManyBody().strength(-600))
      .force('collision', d3.forceCollide().radius(30))
      .force('y', d3.forceY().y(GRAPH_HEIGHT / 2).strength(focusNode ? 0 : 0.1));
    applyLayout(width);

    const link = svg.append('g').attr('class', 'link').selectAll('g').data(edges).enter().append('g');
    link.append('line').attr('class', 'beam').attr('stroke', 'url(#aurora-gradient)');
    link.append('line').attr('class', 'beam-flow');
    link.append('text').attr('dy', -6).text((d) => d.type);

    const tooltip = ensureTooltip();
    const node = svg.append('g').attr('class', 'node').selectAll('g').data(nodes).enter().append('g')
      .call(dragBehaviour(simulation))
      .on('mouseover', (event, d) => {
        tooltip.html(tooltipHtmlOf(d)).transition().duration(200).style('opacity', 0.95)
          .style('left', Math.min(event.pageX + 10, window.innerWidth - 220) + 'px')
          .style('top', (event.pageY - 28) + 'px');
      })
      .on('mouseout', () => tooltip.transition().duration(500).style('opacity', 0));

    appendCelestialBody(node, focusNode);
    node.on('click', (event, d) => {
      if (event.defaultPrevented) return;
      window.location.href = `/static/person_view.html?id=${encodeURIComponent(d.id)}`;
    });
    node.append('text').attr('class', 'name').attr('dx', AVATAR_RADIUS + 6).attr('dy', '.35em').text((d) => d.name);

    simulation.on('tick', () => {
      link.selectAll('line')
        .attr('x1', (d) => d.source.x).attr('y1', (d) => d.source.y)
        .attr('x2', (d) => d.target.x).attr('y2', (d) => d.target.y);
      link.select('text')
        .attr('x', (d) => (d.source.x + d.target.x) / 2)
        .attr('y', (d) => (d.source.y + d.target.y) / 2 - 6);
      node.attr('transform', (d) => `translate(${d.x},${d.y})`);
    });

    const onResize = () => {
      applyLayout(container.node().getBoundingClientRect().width);
      simulation.alpha(0.3).restart();
    };
    window.addEventListener('resize', onResize);
    stopPrevious = () => {
      simulation.stop();
      window.removeEventListener('resize', onResize);
    };

    return edges.length === graph.edges.length ? '' : 'Niektóre relacje są nieprawidłowe i zostały pominięte.';
  }

  window.renderRelationsGraph = renderRelationsGraph;
  window.planetPaletteOf = planetPaletteOf;
  window.initialsOf = initialsOf;
})();
