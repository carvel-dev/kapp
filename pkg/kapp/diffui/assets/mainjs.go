package assets

const (
	mainJS = `
function DependencyGraph($graph, nodes, links) {
  var changeByIdx = {};
  var changeChildren = {};

  for (var i in nodes) {
    changeByIdx[i] = nodes[i];
  }

  for (var i in links) {
    if (!changeChildren[links[i].source]) {
      changeChildren[links[i].source] = [];
    }
    changeChildren[links[i].source].push(links[i].target);
  }

  for (var i in nodes) {
    $graph.append(buildChange(i));
  }

  $graph.on("click", "a", function() {
    var $change = $(this).closest("li");
    if ($change.attr("data-expanded") != "true") {
      $change.attr("data-expanded", "true");
      var changeIdx = $change.attr("data-change-idx");
      $change.append(expandWaitingFor(parseInt(changeIdx)));
    } else {
      $change.removeAttr("data-expanded");
      $("ul", $change).remove();
    }
    return false;
  });

  function expandWaitingFor(changeIdx) {
    var $children = $("<ul/>");
    var children = changeChildren[changeIdx] || [];
    for (var j in children) {
      $children.append(buildChange(children[j]));
    }
    return $children;
  }

  function buildChange(changeIdx) {
    var $change = $("<li/>").attr("data-change-idx", changeIdx+"");
    var change = nodes[changeIdx];
    var children = changeChildren[changeIdx] || [];

    if (children.length > 0) {
      $change.append("<span>"+change.id+" <a href=''>(+"+children.length+")</a></span>");
    } else {
      $change.append("<span>"+change.id+" (0)</span>");
    }
    return $change;
  }

  return {};
}

$(document).ready(function() {
  DependencyGraph($("#deps"), window.diffData.nodes, window.diffData.links);
});
`
)
