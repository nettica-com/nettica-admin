<template>
  <div ref="container" class="d3-network-container">
    <svg ref="svg" :width="options.size.w" :height="options.size.h"></svg>
  </div>
</template>

<script setup>
import { ref, watch, onMounted, onBeforeUnmount, nextTick } from 'vue'
import * as d3 from 'd3'

const props = defineProps({
  netNodes: { type: Array, default: () => [] },
  netLinks: { type: Array, default: () => [] },
  options: {
    type: Object,
    default: () => ({
      force: 3000,
      size: { w: 800, h: 300 },
      nodeSize: 20,
      nodeLabels: true,
      linkLabels: false,
    }),
  },
})

const container = ref(null)
const svg = ref(null)
let simulation = null

const nodeRadius = 10

async function buildGraph() {
  await nextTick()
  if (!svg.value || !props.netNodes.length) return

  const nodes = props.netNodes.map((n) => ({ ...n }))
  const links = props.netLinks.map((l) => ({ source: l.sid, target: l.tid, _color: l._color }))

  if (simulation) simulation.stop()

  const svgEl = d3.select(svg.value)
  svgEl.selectAll('*').remove()

  const linksGroup = svgEl.append('g')
  const nodesGroup = svgEl.append('g')

  const linkEls = linksGroup
    .selectAll('line')
    .data(links)
    .join('line')
    .attr('stroke', (d) => d._color || '#fff')
    .attr('stroke-width', 1.5)
    .attr('stroke-opacity', 0.7)

  const nodeEls = nodesGroup
    .selectAll('g.node-group')
    .data(nodes)
    .join('g')
    .attr('class', 'node-group')

  nodeEls
    .append('circle')
    .attr('r', nodeRadius)
    .attr('fill', (d) => d._color || '#336699')
    .attr('stroke', '#5b81a7')
    .attr('stroke-width', 1.5)

  if (props.options.nodeLabels) {
    nodeEls
      .append('text')
      .attr('dy', '-12')
      .attr('text-anchor', 'middle')
      .attr('font-size', 11)
      .attr('fill', 'white')
      .text((d) => d.name)
  }

  simulation = d3
    .forceSimulation(nodes)
    .force('charge', d3.forceManyBody().strength(-(props.options.force || 3000) / 10))
    .force(
      'link',
      d3.forceLink(links).id((d) => d.id).distance(80),
    )
    .force('center', d3.forceCenter(props.options.size.w / 2, props.options.size.h / 2))
    .force('collision', d3.forceCollide(nodeRadius + 5))

  simulation.on('tick', () => {
    linkEls
      .attr('x1', (d) => d.source.x)
      .attr('y1', (d) => d.source.y)
      .attr('x2', (d) => d.target.x)
      .attr('y2', (d) => d.target.y)

    nodeEls.attr('transform', (d) => `translate(${d.x},${d.y})`)
  })

  nodeEls.call(
    d3
      .drag()
      .on('start', (event, d) => {
        if (!event.active) simulation.alphaTarget(0.3).restart()
        d.fx = d.x
        d.fy = d.y
      })
      .on('drag', (event, d) => {
        d.fx = event.x
        d.fy = event.y
      })
      .on('end', (event, d) => {
        if (!event.active) simulation.alphaTarget(0)
        d.fx = null
        d.fy = null
      }),
  )
}

onMounted(buildGraph)

watch(
  () => [props.netNodes, props.netLinks],
  () => buildGraph(),
  { deep: true },
)

onBeforeUnmount(() => {
  if (simulation) simulation.stop()
})
</script>

<style scoped>
.d3-network-container {
  display: flex;
  justify-content: center;
}
.node-group {
  cursor: grab;
}
</style>
