// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package assets

const (
	mainJS = `
function DependencyGraph($graph, allChanges, linearizedChangeSections, blockedChanges) {
  var changeByID = {};

  for (var i in allChanges) {
    changeByID[allChanges[i].id] = allChanges[i];
  }

  for (var i in linearizedChangeSections) {
    $subgraph = $("<ol></ol>");
    for (var j in linearizedChangeSections[i]) {
      $subgraph.append(buildChange(linearizedChangeSections[i][j]));
    }
    $graph.append($("<li class=\"section\"></li>").append($subgraph));
  }
  if (blockedChanges.length > 0) {
    $subgraph = $("<ol></ol>");
    for (var i in blockedChanges) {
      $subgraph.append(buildChange(blockedChanges[i]));
    }
    $graph.append($("<li class=\"section blocked\"><p>Blocked changes</p></li>").append($subgraph));
  }

  $graph.on("click", "a.expand", function() {
    var $change = $(this).closest("li");
    if ($change.attr("data-expanded") != "true") {
      $change.attr("data-expanded", "true");
      var changeID = $change.attr("data-change-id");
      $change.append(expandWaitingFor(changeID));
    } else {
      $change.removeAttr("data-expanded");
      $("ul", $change).remove();
    }
    return false;
  });

  $graph.on("click", "a.highlight", function() {
    var $change = $(this).closest("li");

    if ($change.hasClass("highlighted-for")) {
      $change.removeClass("highlighted-for");
      $graph.removeClass("highlighted");
    } else {
      // remove old highlight
      $(".highlighted-for", $graph).removeClass("highlighted-for");
      $(".highlighted", $graph).removeClass("highlighted");

      var changeID = $change.attr("data-change-id");
      var children = changeByID[changeID].waitingForIDs || [];
      for (var i in children) {
        $("li[data-change-id=\""+children[i]+"\"]", $graph).addClass("highlighted");
      }

      // show highlighting
      $change.addClass("highlighted-for");
      $graph.addClass("highlighted");
    }
    return false;
  });

  function expandWaitingFor(changeID) {
    var $children = $("<ul/>");
    var children = changeByID[changeID].waitingForIDs || [];
    for (var i in children) {
      $children.append(buildChange(children[i]));
    }
    return $children;
  }

  function buildChange(changeID) {
    var $change = $("<li/>").attr("data-change-id", changeID);
    var change = changeByID[changeID];
    var children = change.waitingForIDs || [];

    if (children.length > 0) {
      $change.append("<span>"+change.name+" "+
        "<a href='' class='expand'>(+"+children.length+")</a> "+
        "<a href='' class='highlight'>^</a></span>");
    } else {
      $change.append("<span>"+change.name+" (0)</span>");
    }
    return $change;
  }

  return {};
}

$(document).ready(function() {
  DependencyGraph($("#deps"),
    window.diffData.allChanges || [],
    window.diffData.linearizedChangeSections || [],
    window.diffData.blockedChanges || [],
  );
});
`
)
