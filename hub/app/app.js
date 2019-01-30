
var margin = {top: 1, right: 1, bottom: 6, left: 1},
    width = 1060 - margin.left - margin.right,
    height = 500 - margin.top - margin.bottom;

var formatNumber = d3.format(",.0f"),
    format = function(d) { return formatNumber(d / 10) + " Packets PS last 10 seconds"; },
    color = d3.scale.category20();

var svg = d3.select("svg")
    .attr("width", width + margin.left + margin.right)
    .attr("height", height + margin.top + margin.bottom)
  .append("g")
    .attr("transform", "translate(" + margin.left + "," + margin.top + ")");

var sankey = d3.sankey()
    .nodeWidth(15)
    .nodePadding(10)
    .size([width, height]);

var path = sankey.link();
var loaded = false
var freqCounter = 1;

var graph = {
  nodes : [],
  links : []	
};
var particles = [];
var elapsed = 1


// Create WebSocket connection.
var ip = window.location.hostname
const socket = new WebSocket('ws://'+ip+':9999/data');

// Connection opened
socket.addEventListener('open', function (event) {
   // socket.send('Hello Server!');
	console.log("opened server")
});


//d3.json("data2.json", function(graph) {
// Listen for messages
socket.addEventListener('message', function (event) {
    console.log('Message from server ', event.data);

     graph = JSON.parse(event.data);
     sankey
      .nodes(graph.nodes)
      .links(graph.links)
      .layout(32).relayout();

svg.selectAll("*").remove();
var context = d3.select("canvas").node().getContext("2d")
context.clearRect(0,0,1000,1000);

loaded = true
var link = svg.append("g").selectAll(".link")
      .data(graph.links)
      .enter().append("path")
      .attr("class", "link")
      .attr("d", path)
      .style("stroke-width", function(d) { return Math.max(1, d.dy); })
      .sort(function(a, b) { return b.dy - a.dy; });

  link.append("title")
      .text(function(d) { return d.source.name + " → " + d.target.name + "\n" + format(d.value); });

  var node = svg.append("g").selectAll(".node")
      .data(graph.nodes)
    .enter().append("g")
      .attr("class", "node")
      .attr("transform", function(d) { return "translate(" + d.x + "," + d.y + ")"; })
    .call(d3.behavior.drag()
      .origin(function(d) { return d; })
      .on("dragstart", function() { this.parentNode.appendChild(this); })
      .on("drag", dragmove));

  node.append("rect")
      .attr("height", function(d) { return d.dy; })
      .attr("width", sankey.nodeWidth())
      .style("fill", function(d) { return d.color = color(d.name.replace(/ .*/, "")); })
      .style("stroke", "none")
    .append("title")
      .text(function(d) { return d.name + "\n" + format(d.value); });

  node.append("text")
      .attr("x", -6)
      .attr("y", function(d) { return d.dy / 2; })
      .attr("dy", ".35em")
      .attr("text-anchor", "end")
      .attr("transform", null)
      .text(function(d) { return d.name; })
    .filter(function(d) { return d.x < width / 2; })
      .attr("x", 6 + sankey.nodeWidth())
      .attr("text-anchor", "start");

  function dragmove(d) {
    d3.select(this).attr("transform", "translate(" + d.x + "," + (d.y = Math.max(0, Math.min(height - d.dy, d3.event.y))) + ")");
    sankey.relayout();
    link.attr("d", path);
  }

  var linkExtent = d3.extent(graph.links, function (d) {return d.value});
  var frequencyScale = d3.scale.linear().domain(linkExtent).range([0.05,1]);
  var particleSize = d3.scale.linear().domain(linkExtent).range([1,5]);


  graph.links.forEach(function (link) {
    link.freq = frequencyScale(link.value);
    link.particleSize = 2.5;
    link.particleColor = d3.scale.linear().domain([0,1])
    .range([link.source.color, link.target.color]);
  })

  var t = d3.timer(tick, 1000);
  var particles = [];

  function tick(elapsed, time) {

    particles = particles.filter(function (d) {return d.current < d.path.getTotalLength()});

    d3.selectAll("path.link")
    .each(
      function (d) {
//        if (d.freq < 1) {
          var offset = (Math.random() - .5) * d.dy;
          if (Math.random() < d.freq) {
            var length = this.getTotalLength();
            particles.push({link: d, time: elapsed, offset: offset, path: this, length: length, animateTime: length})
          }
//        }
/*        else {
          for (var x = 0; x<d.freq; x++) {
            var offset = (Math.random() - .5) * d.dy;
            particles.push({link: d, time: elapsed, offset: offset, path: this})
          }
        } */
      });

    particleEdgeCanvasPath(elapsed);
  }



  function particleEdgeCanvasPath(elapsed) {
    var context = d3.select("canvas").node().getContext("2d")

    context.clearRect(0,0,1000,1000);

      context.fillStyle = "gray";
      context.lineWidth = "1px";
    for (var x in particles) {
        var currentTime = elapsed - particles[x].time;
        particles[x].current = currentTime * 0.15;
        var currentPos = particles[x].path.getPointAtLength(particles[x].current);
        context.beginPath();
      context.fillStyle = particles[x].link.particleColor(particles[x].current/particles[x].path.getTotalLength());
        context.arc(currentPos.x,currentPos.y + particles[x].offset,particles[x].link.particleSize,0,2*Math.PI);
        context.fill();
    }
  }
  	console.log('complete round')
});
