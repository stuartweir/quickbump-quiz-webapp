// Graph object thing used to render answer data for questions ...
// Construct it by passing a question
//     new Graph(domElement, myQuestion)
// Call load, passing an answer collection, to update it
//     myGraph.load({'answerId': {..., Response: [...]}})
// Call reset to reset it back to having no answers

var Graph = function(elemid, question) {
    this._Q = question;
    this._chart = d3.select(elemid).append('div')
        .attr('class', 'graph')
        ;
    //this._maxwidth = this._chart.property('clientWidth');

    this.reset = function() {
        var choices = this._Q.Data.Info.Choices;
        var data = this._data = [];
        for (var i=0; i<choices.length; i++) {
            data[i] = {key: i, text: choices[i], count: 0, w: '0%'};
        }
    }

    this.reset();
    var line = this._chart.selectAll('div').data(this._data, function(e) {
        return e.key;
    }).enter().append('div')
        .attr('class', 'line');
    line.append('div')
        .attr('class', 'graph-label');
    line.append('div').attr('class', 'barwrapper')
        .append('div')
        .attr('class', 'bar')
        .style('width', 0)
        .style('background-color', colorgenerator())
        ;

    this.load = function(answers) {
        var data = this._data;
        // Update data
        for (id in answers) {
            var a = answers[id];
            if(a.Response) a.Response.map(function(x) {
                data[x].count++;
            });
        }
        // Update UI
        var max = Math.max.apply(Math, data.map(function(e){ return e.count; })) || 1;
        var line = this._chart.selectAll('.line').data(this._data, function(e){ return e.key; });
        line.select('.graph-label')
            .text(function(d) { return d.text; });
        line.select('.bar')
            .text(function(d) { return d.count; })
            .transition().duration(550)
            .tween('style.width', function(d, i, a) {
                i = d3.interpolate(d.w, d.w=100*d.count/max+'%');
                return function(t) {
                    this.style.setProperty('width', i(t));
                }
            });
    }

    this.load({}) // Draw the labels or whatever
}

function colorgenerator() {
    var colorlist = [
        chameleon1, skyblue1, scarletred1, plum1, orange1
        ];
    var c = 0;
    return function() {
        return colorlist[c++%colorlist.length];
    }
}
