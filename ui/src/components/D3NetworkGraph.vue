<template>
  <div ref="container" class="d3-network-container">
    <svg ref="svg" :width="options.size.w" :height="options.size.h">
      <g ref="linksGroup">
        <line
          v-for="(link, i) in renderedLinks"
          :key="'link-' + i"
          :stroke="link._color || '#fff'"
          stroke-width="1.5"
          stroke-opacity="0.7"
        />
      </g>
      <g ref="nodesGroup">
        <g
          v-for="node in renderedNodes"
          :key="'node-' + node.id"
          class="node-group"
        >
          <circle
            :r="nodeRadius"
            :fill="node._color || '#336699'"
            stroke="#5b81a7"
            stroke-width="1.5"
          />
          <text
            v-if="options.nodeLabels"
            dy="-12"
            text-anchor="middle"
            font-size="11"
            fill="white"
          >{{ node.name }}</text>
        </g>
      </g>
    </svg>
  </div>
</template>

<script setup>
import { ref, watch, onMounted, onBeforeUnmount } from 'vue'
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
const linksGroup = ref(null)
const nodesGroup = ref(null)
const renderedNodes = ref([])
const renderedLinks = ref([])
let simulation = null

const nodeRadius = 10

function buildGraph() {
  if (!svg.value || !props.netNodes.length) return

  const nodes = props.netNodes.map((n) => ({ ...n }))
  const links = props.netLinks.map((l) => ({ source: l.sid, target: l.tid, _color: l._color }))

  renderedNodes.value = nodes
  renderedLinks.value = props.netLinks

  if (simulation) simulation.stop()

  simulation = d3
    .forceSimulation(nodes)
    .force('charge', d3.forceManyBody().strength(-(props.options.force || 3000) / 10))
    .force(
      'link',
      d3
        .forceLink(links)
        .id((d) => d.id)
        .distance(80),
    )
    .force(
      'center',
      d3.forceCenter(props.options.size.w / 2, props.options.size.h / 2),
    )
    .force('collision', d3.forceCollide(nodeRadius + 5))

  const svgEl = d3.select(svg.value)
  const linkEls = d3.select(linksGroup.value).selectAll('line').data(links)
  const nodeEls = d3.select(nodesGroup.value).selectAll('g.node-group').data(nodes)

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

  // suppress unused var warning — svgEl used for potential zoom extension
  void svgEl
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
