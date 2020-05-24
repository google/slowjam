/*
Copyright 2020 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import "text/template"

var fmap = template.FuncMap{
	"Milliseconds": milliseconds,
	"Creator":      creator,
	"Height":       height,
	"Color":        callColor,
}

// reportTemplate is the HTML template for the output report, to avoid compile issues embeding in the code
var reportTemplate = template.Must(template.New("SlowJamReportTemplate").Funcs(fmap).Parse(`<html>
<head>
  <script type="text/javascript" src="https://www.gstatic.com/charts/loader.js"></script>
  <script type="text/javascript">
	google.charts.load('current', {'packages':['timeline']});

	google.charts.setOnLoadCallback(drawTimeline);
	function drawTimeline() {
	  var container = document.getElementById('timeline');
	  var chart = new google.visualization.Timeline(container);
	  var dataTable = new google.visualization.DataTable();
	  dataTable.addColumn({ type: 'string', id: 'Layer' });
	  dataTable.addColumn({ type: 'string', id: 'Function' });
	  dataTable.addColumn({ type: 'string', id: 'style', role: 'style' });
	  dataTable.addColumn({ type: 'date', id: 'Start' });
	  dataTable.addColumn({ type: 'date', id: 'End' });

	  dataTable.addRows([
		{{ range $g := .TL.Goroutines }}
		  {{ range $index, $layer := .Layers}}
			{{ range $layer.Calls }}
			  [ '{{ $g.ID }}: {{ $g.Signature | Creator }}', '{{ .Name }}', '{{ Color .Package $index }}',  new Date({{ .StartDelta | Milliseconds }}), new Date({{ .EndDelta | Milliseconds }}) ],
			{{ end }}
		  {{ end }}
		{{ end }}
	  ]);
	  var options = {
		avoidOverlappingGridLines: false,
	  };
	  chart.draw(dataTable, options);
	}
  </script>
</head>
<body>
  <h1>SlowJam for {{ .Duration}} ({{ .TL.Samples }} samples) - <a href="/slowjam_full.html">full</a> | <a href="/slowjam_simple.html">simple</a></h1>
  <div id="timeline" style="width: 3200px; height: 1024px;"></div>
</body>
</html>`))
