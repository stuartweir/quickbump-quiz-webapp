var CheckableChoiceList = ChoiceListBase.extend({
    events: {
        'click      .feature-checkstate':  'toggleCheckState',
    },

    toggleCheckState: function(ev) {
        console.log(ev);
        var itemel = $(ev.currentTarget).closest('.choiceitem');
        var model = itemel.data('view').model;
        model.set('checked', !model.get('checked'));
    },

    itemData: function(item, data) {
        data.checked = item.model.get('checked');
    },

    appendItem: function() {
        var item = ChoiceListBase.prototype.appendItem.apply(this, arguments);
        item.$el.find('input').attr('readonly', '');

        var feature = item.addFeature({
            icon: 'icon-check-empty',
            name: 'checkstate',
        });
        item.model.set('checked', false)
        item.model.on('change:checked', function(model, isChecked) {
            if (isChecked)
                feature.removeClass('icon-check-empty').addClass('icon-check');
            else
                feature.removeClass('icon-check').addClass('icon-check-empty');
        }, this);

        return item;
    },
});

var AnswerQuestionModel = Backbone.Model.extend({
    validate: function(attrs, options) {
        // todo
    },
});

var AnswerQuestionView = QBView.extend({
    events: {
        'submit     form':              'onSubmit',
    },

    noop: function() {
        console.log('adfjasdkfjadsljfadslkfajsklfdjf')
    },

    initialize: function(opts) {
        QBView.prototype.initialize.apply(this, arguments);

        this.error = ElementHelper('errorbox');
        this.textinfo = ElementHelper('answer-textquestioninfo');
        this.choiceinfo = ElementHelper('answer-choicequestioninfo');

        this.choicelist = new CheckableChoiceList({
            el: $('#choicelist'),
            model: new Backbone.Model({
                items: [],
                checked: [],
            }),
        })

        this.model = new AnswerQuestionModel({
            qid: opts.args[0],
            //question: ,
            error: null,
        })
        //.on('all', function() { console.log(this, arguments); })
        .on('change:error', function(model, value) {
            if (value) 
                this.error.$el.html(this.error.tmpl(value)).show();
            else
                this.error.$el.hide().html()
        }, this)
        .on('change:question', this.onNewQuestion, this)
        .on('change:form-error', function(model, err) {
            if (err) {
                this.$el.find('#validation-errorbox').text(err).show();
            } else {
                this.$el.find('#validation-errorbox').hide();
            }
        }, this)
        .on('change', function(model) {
            //var errs = model.validate(model.attributes);
            // todo
        }, this)
        .set('form-error', null)
        ;
        
        this.choicelist.model.on('change:checked', function(choicemodel, value) {
            this.model.set('checked', value);
        }, this);

        this.getQuestion();
        this.onNewQuestion();
    },

    onNewQuestion: function() {
        var question = this.model.get('question');
        if (!question) {
            this.$el.find('form').hide();
        } else {
            this.$el.find('form').show();
            if (question.Data.Mode == ChoiceMode) {
                this.$el.find('.text-only').hide();
                this.$el.find('.choice-only').show();
                this.choicelist.loadItems(question.Data.Info.Choices);
                this.choiceinfo.$el.html(this.choiceinfo.tmpl({
                    qid: this.model.get('qid'),
                    question: question,
                }));
            } else {
                this.$el.find('.text-only').show();
                this.$el.find('.choice-only').hide();
                this.textinfo.$el.html(this.textinfo.tmpl({question: question}));
            }
        }
    },

    getQuestion: function() {
        var model = this.model;
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
    },

    onSubmit: function() {
        var model = this.model;
        model.set('form-error', null);

        var disabled = this.$el.find('button[type=submit]').attr('disabled', '');
        var answer = {};
        answer.QuestionId = this.model.get('qid');
        if (this.model.get('question').Data.Mode == TextMode) {
            // todo, use the model and do validation and stuff ...
            answer.Response = this.$el.find('textarea').val();
        } else {
            var selected = [];
            var checked = this.model.get('checked');
            var choices = this.model.get('question').Data.Info.Choices;
            for (var idx=0; idx < choices.length; idx++) {
                if (checked[idx]) selected.push(idx);
            };
            console.log(selected);
            answer.Response = selected;
        }

        $.post('/api/answer', JSON.stringify(answer), function() {
            // What should we even do here?
            router.viewQuestion(answer.QuestionId);
        }).fail(function(resp) {
            model.set('form-error', 'Submission failed: ' + resp.responseText);
        }).always(function() {
            disabled.removeAttr('disabled');
        });
    },
});
