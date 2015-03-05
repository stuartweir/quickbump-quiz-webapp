var QuestionModes = {
    MultipleChoice: 1,
    ShortAnswer: 2,
    Configurable: 4,
}

QuestionCreateModel = Backbone.Model.extend({
    validate: function(attrs, options) {
        var errs = {}
        var numchoices = attrs.choices.length;

        if (!attrs.Question)
            errs.Question = 'Question cannot be blank';

        ['MinChoices', 'MaxChoices', 'WordLimit', 'CharLimit'].forEach(
            function(attr) {
                if (isNaN(attrs[attr])) {
                    errs[attr] = 'An integer is required';
                }
            });

        if (attrs.MinChoices > attrs.MaxChoices) {
            errs.MinChoices = '';
            errs.MaxChoices = 'Minimum Choices must be less than or equal to Maximum Choices';
        } else {
            if (attrs.MinChoices > numchoices)
                errs.MinChoices = 'Minimum Choices must be less than or equal to the number of choices offered';
            if (attrs.MaxChoices > numchoices)
                errs.MaxChoices = 'Maximum Choices must be less than or equal to the number of choices offered';
        }
        if (Object.keys(errs).length) return errs;
    },
});

QuestionCreateView = QBView.extend({
    inputNamed: function(name) {
        return this.$el.find('[name="'+name+'"]');
    },

    events: {
        'change     input':             'onInputChange',
        'toggled    .mode-btns':        'onModeChange',
        'submit':                       'onSubmit',
    },
    
    render: function() {
        if (this.$el.html() == '') {
            this.$el.html(this.template());
            for (attr in this.model.attributes) {
                this.inputNamed(attr).val(this.model.get(attr));
            }
            this.choicelist = new EditableChoiceList({
                el: this.$el.find('#choicelist'),
                model: new Backbone.Model({
                    items: this.model.get('choices'),
                }),
            });
            this.choicelist.model.on('change:items', function(model, value) {
                this.model.set('choices', value);
            }, this);
            this.showModelError();
        }
        return this;
    },

    onInputChange: function(ev) {
        var target = $(ev.target);
        var val = target.val();
        if (target.attr('type').toLowerCase() == 'number') {
            // rounding is cheating, we just happen to know we want integers
            val = Math.round(val) == val ? Number(val) : NaN;
        }
        this.model.set(target.attr('name'), val);
    },

    onModeChange: function(ev, el) {
        this.model.set('mode', Number(el.val()));
    },

    onSubmit: function(ev) {
        // Disable UI
        var disabled = this.$el.find('#submit-question').attr('disabled', '')
        // Prepare model
        var mode = this.model.get('mode') == QuestionModes.ShortAnswer ?
            TextMode : ChoiceMode;
        var info = {
            Question: this.model.get('Question')
        }
        if (mode == TextMode) {
            info.WordLimit = this.model.get('WordLimit');
            info.CharLimit = this.model.get('CharLimit');
        } else {
            info.MinChoices = this.model.get('MinChoices');
            info.MaxChoices = this.model.get('MaxChoices');
            info.Choices = this.model.get('choices');
        }
        var question = {
            End: new Date(),
            Data: {
                Mode: mode,
                Info: info,
            },
        }
        var errbox = this.$el.find('#errorbox');
        $.post('/api/question', JSON.stringify(question), function(qid) {
            router.viewQuestion(qid)
        }).fail(function(resp) {
            errbox.html('Submission failed. ' + resp.responseText).show();
        }).always(function() {
            // Reeanble UI
            disabled.removeAttr('disabled');
        });
        return false;
    },

    showModelError: function(errors) {
        var errel = this.$el.find('#errorbox');
        this.$el.find('input').removeClass('invalid');
        if (errors) {
            var msgs = []
            for (name in errors) {
                this.inputNamed(name).addClass('invalid');
                if (errors[name] && errors[name].length)
                    msgs.push(errors[name]);
            }
            errel.html('<ul><li>' + msgs.join('</li><li>') + '</li></ul>').show();
        } else {
            errel.hide();
        }
    },

    initialize: function(opts) {
        this.model = new QuestionCreateModel({
            mode: null,
            Question: 'What... is the air-speed velocity of an unladen swallow?',
            choices: [''],
            MaxChoices: 1,
            MinChoices: 1,
            WordLimit: 0,
            CharLimit: 0,
        }),
        this.template = opts.template;

        var model = this.model;
        
        this.model.on('change:mode', function() {
            var mode = model.get('mode');
            switch (mode) {
            case QuestionModes.MultipleChoice:
                this.inputNamed('MinChoices').val(1);
                this.inputNamed('MaxChoices').val(1);
                model.set('MaxChoices', 1);
                model.set('MinChoices', 1);
            case QuestionModes.Configurable:
                this.$el.find('.choice-only').show();
                this.$el.find('.text-only').hide();
                if (mode == QuestionModes.Configurable) {
                    this.inputNamed('MinChoices').removeAttr('disabled');
                    this.inputNamed('MaxChoices').removeAttr('disabled');
                } else {
                    this.inputNamed('MinChoices').attr('disabled', '');
                    this.inputNamed('MaxChoices').attr('disabled', '');
                }
                break;
            case QuestionModes.ShortAnswer:
                this.$el.find('.choice-only').hide();
                this.$el.find('.text-only').show();
                break;
            }
        }, this);

        this.model.on('change', function(model) {
            var errs = model.validate(model.attributes);
            this.showModelError(errs);
            if (errs || model.get('choices').length == 0) {
                this.$el.find('#submit-question').attr('disabled', '');
            } else {
                this.$el.find('#submit-question').removeAttr('disabled');
            }
        }, this);

        this.render();
        this.model.set('mode', QuestionModes.MultipleChoice);
    },
});
