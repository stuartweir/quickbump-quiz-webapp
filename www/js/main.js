var router = null;

var LandingView = QBView.extend({
    initialize: function(opts) {
        var view = this;
        this.template = opts.template;
        this.render();
    },
});

var MissingPageView = QBView.extend({});

$(document).ready(function() {
    router = (new (Backbone.Router.extend({
        routes: {
            '':           'landing',
            'new':        'newquestion',
            'view/:id':   'viewquestion',
            'answer/:id': 'answerquestion',
            '*path':      '404',
        },
        _uninstallView: function() {
            if (this._current_view) {
                this._current_view.remove();
                this._current_view = null;
            }
        },
        _installView: function(view) {
            if (this._current_view) console.warn('What are you doing?')
            this._current_view = view;
        },
        viewQuestion: function(qid) {
            this.navigate('view/'+qid, {trigger: true});
        },
        answerQuestion: function(qid) {
            this.navigate('answer/'+qid, {trigger: true});
        },
    })))
    //.on('all', function(){ console.log(this, arguments); }) // dbg
    ;
 
    // This function takes an object about 20 lines below and sets up the
    // router to construct the view (right hand side) whenever the paired
    // event is fired
    (function(obj) {
        Object.keys(obj).map(function(name) {
            router.on('route:' + name, function() {
                var cls = obj[name];
                this._uninstallView();
                var $el = $(document.createElement('div')).attr('id', name).show();
                document.body.appendChild($el[0]);
                this._installView(new cls({
                    el: $el,
                    template: get_template(name),
                    args: arguments, // oh dear ...
                }));
            });
        });
    })({
        'landing': LandingView,
        'newquestion': QuestionCreateView,
        'viewquestion': QuestionView,
        'answerquestion': AnswerQuestionView,
        '404': MissingPageView,
    });

    Backbone.history.start(); // ... Is this enough magic for you?

    new (Backbone.View.extend({
        events: {
            'click      #post-an-answer-btn':   'doAnswer',
        },
        doAnswer: function() {
            var editor = this.$el.find('[name=target-question]');
            var qid = editor.val();
            if(!qid.length) {
                alert('Question Identifier required');
            } else {
                editor.val('')
                router.answerQuestion(qid);
            }
        },
        initialize: function() {
            console.log('hew');
        },
    }))({
        el: $('#navheader'),
    });
});
