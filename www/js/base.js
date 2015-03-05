var TextMode = 'TEXT';
var ChoiceMode = 'CHOICE';

function format_date(datestring) {
    var date = new Date(datestring);
    return date.toLocaleString()
}

function get_template(id) {
    var tmplstring = $('script[type="text/underscore-template"][for="'+id+'"]').html();
    if (!tmplstring) {
        console.error('No template found for #'+id);
        return function(){/* noop */};
    }
    return _.template(tmplstring);
}

function setVisible(target, value) {
    return value ? target.show() : target.hide();
}

function ElementHelper(id) {
    var el = $('#'+id);
    return {
        el: el[0],
        $el: el,
        tmpl: get_template(id),
    }
}

var QBView = Backbone.View.extend({
    initialize: function(opts) {
        this.template = opts.template;
        this.render();
    },
    render: function() {
        if (this.$el.html() == '')
            this.$el.html(this.template());
        return this;
    },
});

// bootstrap buttons don't emit any events when they are toggled ...
var _bootstrap_button_toggle = $.fn.button.Constructor.prototype.toggle
$.fn.button.Constructor.prototype.toggle = function() {
    _bootstrap_button_toggle.apply(this, arguments);
    var $parent = this.$element.closest('[data-toggle="buttons-radio"]');
    $parent && $parent.trigger('toggled', [$parent.find('.active')]);
}

$(document).on('click', 'a.modal-image-display', function(ev) {
    var target = $(ev.currentTarget);
    var href = target.attr('href');
    var title = target.attr('alt');
    var dlg = $(document.createElement('div')).addClass('modal').html("\
        <div class='modal-header'>\
            " + title + "\
        </div>\
        <div class='modal-body'>\
            <center>\
                <img src='" + href + "' />\
            </center>\
        </div>\
        <div class='modal-footer'>\
            <button class='btn' data-dismiss='modal' aria-hidden='true'>Close</button>\
        </div>\
        ");
    dlg.modal();
    dlg.on('hidden', function() {
        dlg.remove();
    });
    dlg.on('shown', function() {
        dlg.find('button').focus();
    });
    return false;
});
