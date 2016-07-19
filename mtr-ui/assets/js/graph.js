var msOffset = new Date().getTimezoneOffset()*60*1000; // offset in milliseconds

function utcToLocal(x) {
	return Date.parse(x) - msOffset;
}

function plotData(div, csvData, graphOptions, thresholds) {
    //Grey box for thresholds if specified:
    if (thresholds) {
        graphOptions.underlayCallback = function(canvas, area, g) {
          // convert coords to Dom x/y.  Only concerned with Y values
          var bottom = g.toDomCoords(0, thresholds[0]);
          var top = g.toDomCoords(0, thresholds[1]);

          canvas.fillStyle = "rgba(200, 200, 200, 0.5)";
          //start at (0,0) which is upper left, and make a box from bottom left to top right
          canvas.fillRect(0, top[1], area.w + area.x, bottom[1]-top[1]);
        }
    }

//    // Make the default time plotting in local time
//    if (!graphOptions.xValueParser) {
//        graphOptions.xValueParser = utcToLocal;
//    }

    var g = new Dygraph(
        div,
        csvData,
        graphOptions);
}

function showGraph(csvUrl, res, graphOptions, thresholds) {
    var graphElement = document.getElementById('graphdiv');
    var div = document.createElement('div');
    div.style.width = '90vw'; // use 90% of the available width (scales with changing width)
    div.style.height = '40vh';
    div.style.display = 'inline-block';
    div.style.margin = '4px';
    // appending to parent div lets us plots as many graphs as we like
    graphElement.appendChild(div);

    if (!res) {
        res = "minute";
    }
    csvUrl += "&resolution=" + res;

    var request = new XMLHttpRequest();
    request.open('GET', csvUrl, true);
    request.setRequestHeader ("Accept", "text/csv");

    request.onload = function() {
        if (request.status == 200) {
            // Success but check for null data
            var data = request.response;
            if (!data) {
                data = "\n";
                graphOptions.title += " [Error: No Data Found]";
            }

            console.log("DEBUG", div, data, graphOptions, thresholds);
            plotData(div, data, graphOptions, thresholds);
        } else {
            // We reached our target server, but it returned an error.  Plot a blank graph with the error message.
            //throw "error loading csv";
            graphOptions.title += " [Error fetching data: " + request.statusText + "]";
            plotData(div, "\n", graphOptions, thresholds);
        }
    };

    request.onerror = function() {
        // There was a connection error of some sort
        //throw "error downloading CSV data";
        graphOptions.title += " [Error fetching data: " + request.statusText + "]";
        plotData(div, "\n", graphOptions, thresholds);
    };

    request.send();
}