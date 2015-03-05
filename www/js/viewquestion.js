QuestionPageModel = Backbone.Model.extend({});

QuestionView = QBView.extend({
    initialize: function(opts) {
        var view = this;
        var model = this.model = new QuestionPageModel({
            qid: opts.args[0],
            error: null,
            question: null,
            answers: null,
        });
        this.$el.html(opts.template());

        this.graphel = $('#graphcontainer');
        this.graphel.html(''); // fixme, this should happen in graph.reset
        this.graph = null;

        this.error = ElementHelper('errorbox');
        this.questioninfo = ElementHelper('questioninfo');
        this.answerheader = ElementHelper('answerheader');
        this.answerlist = ElementHelper('answerlist');
        this.answerelements = {};

        model.on('change:error', function(){ view.render({error: model.get('error')}); })

        model.on('change:question', function(){ view.render({question: model.get('question')}); })
        model.on('change:question', function(){ view.resetGraph(); })
        model.on('change:question', function(){ view.beginPollingForAnswers(); })

        model.on('change:answers', function(){ view.render({answers: model.get('answers')}); });

        $.get('/api/question/'+model.get('qid'), function(data) {
            model.set('error', null);
            model.set('question', JSON.parse(data));
        }).fail(function(resp) {
            model.set('error', {
                code: resp.status,
                message: resp.responseText,
            });
            model.set('question', null);
        });

        this.render();
    },

    remove: function() {
        if (this._poller) clearTimeout(this._poller);
        return QBView.prototype.remove.apply(this, arguments);
    },

    resetGraph: function() {
        var question = this.model.get('question');
        if (question.Data.Mode == ChoiceMode)
            this.graph = new Graph('#'+this.graphel.attr('id'), question);
    },

    beginPollingForAnswers: function() {
        this.pollForAnswers();
    },

    pollForAnswers: function() {
        if (this.el.parentElement == null) // We have been slain
            return;
        var view = this;
        var model = this.model;
        $.get('/api/answer?Question='+this.model.get('qid'), function(data) {
            data = JSON.parse(data)
            var answers = model.get('answers') || {};
            var newstuff = {};
            for (aid in data) {
                if (answers[aid] == undefined) // a new answer!
                    answers[aid] = newstuff[aid] = data[aid]
            }
            // backbone is stupid and returns a reference to answers instead
            // of a copy, except for when we first set it to an object, our
            // modifications so we have already happened in the model, so it
            // won't realize that any change happened, so we have to do its
            // work for it ...
            if (model.get('answers') != answers) {
                model.set('answers', answers)
                if (view.graph) view.graph.load(answers);
            } else if (Object.keys(newstuff).length != 0) {
                if (view.graph) view.graph.load(newstuff);
                model.trigger('change');
                model.trigger('change:answers');
            }
        }).always(function(){
            view._poller = setTimeout(function() { view.pollForAnswers() }, 3000);
        });
    },
    
    render: function(changes) {
        var isErrored = this.model.get('error') != null;

        isErrored ? this.error.$el.show()        : this.error.$el.hide();
        isErrored ? this.questioninfo.$el.hide() : this.questioninfo.$el.show()
        isErrored ? this.graphel.hide()          : this.graphel.show();
        isErrored ? this.answerlist.$el.hide()   : this.answerlist.$el.show();

        if(changes) {
            if(changes.error)
                this.error.$el.html(this.error.tmpl(changes.error));
            if(changes.question)
                this.questioninfo.$el.html(this.questioninfo.tmpl({
                    qid: this.model.get('qid'),
                    question: changes.question,
                }));
            if(changes.answers)
                this.renderAnswers(changes.answers);
        }
    },

    renderAnswers: function(answers) {
        this.answerheader.$el.html(this.answerheader.tmpl({answers: answers}))
        var new_answer_elems = [];
        for (aid in answers) {
            var el = this.answerelements[aid];
            if (!el) {
                var elcontent = this.answerlist.tmpl({
                    question: this.model.get('question'),
                    answer: answers[aid],
                }).trim();
                var el = this.answerelements[aid] = document.createElement('div');
                el.innerHTML = elcontent; 
                new_answer_elems.push(el);
            }
        }
        if (new_answer_elems.length) {
            var frag = document.createDocumentFragment();
            new_answer_elems.forEach(function(el) {
                frag.insertBefore(el, frag.firstChild);
            });
            this.answerlist.el.insertBefore(frag, this.answerlist.el.firstChild);
        }
        return this;
    },
})

