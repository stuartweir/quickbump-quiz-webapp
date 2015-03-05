// Features are things which are represented by buttons next to items
// They have:
//      icon, strings pointing to classes in FontAwesome
//      name, some presumably unique identifier thing which is prefixed with
//            feature- and set as the button's class
// And that's it, an element is returned by ChoiceItem.addFeature, listen 
// to its events
var ChoiceItem = QBView.extend({
    events: {
        'change     .value-edit':       'onChange',
    },
    onChange: function(ev) {
        this._inChange = true;
        this.model.set('value', ev.currentTarget.value);
    },
    focus: function() {
        this.$el.find('.value-edit').focus();
    },
    addFeature: function(feature) {
        var btn = $(document.createElement('button'))
            .attr('type', 'button')
            .addClass('btn')
            .addClass(feature.icon)
            .addClass('feature-' + feature.name);
        this.$el.append(btn);
        return btn;
    },
    initialize: function() {
        this.template = get_template('choiceitem');
        this.$el.addClass('input-append');
        this.$el.html(this.template());

        this.model.on('change:value', function(model, value) {
            if (this._inChange == true) return;
            this.$el.find('.value-edit').val(value);
        }, this);
    },
});

var ChoiceListBase = QBView.extend({
    /* TODO, document this function and its relationship to itemData
     */
    recomputeModel: function() {
        var sum = {};
        var view = this;
        this.$list.children().map(function(idx, el) {
            var data = {};
            view.itemData($(el).data('view'), data);
            for (name in data)
                (sum[name] == undefined ? sum[name] = [] : sum[name])[idx] = data[name];
        });
        for (name in sum)
            this.model.set(name, sum[name]);
    },

    itemData: function(item, data) {
        data.items = item.model.get('value')
    },

    appendItem: function() {
        var itemel = $(document.createElement('div')).addClass('choiceitem');
        this.$list.append(itemel);

        var item = new ChoiceItem({
            el: itemel,
            model: new Backbone.Model({
                value: '',
            }),
        });
        itemel.data('view', item);
        item.model.on('change', this.recomputeModel, this);
        this.recomputeModel();
        return item;
    },

    setupDOM: function() {
        this.$list = this.$el;
    },

    loadItems: function(items) {
        items.map(function(value) {
            var item = this.appendItem();
            item.model.set('value', value);
        }, this);
    },

    initialize: function() {
        this.$el.html('');
        this.setupDOM(); // should set this.$list
        if (items = this.model.get('items')) this.loadItems(items);
    },
});

// For use in create question
var EditableChoiceList = ChoiceListBase.extend({
    events: {
        'click      .append-item':          'focusAppendItem',
        'click      .feature-remove-item':  'removeItem',
    },

    focusAppendItem: function() {
        this.appendItem().focus();
    },

    appendItem: function() {
        var item = ChoiceListBase.prototype.appendItem.apply(this, arguments);
        item.addFeature({
            icon: 'icon-remove',
            name: 'remove-item',
        });
        return item;
    },

    removeItem: function() {
        var itemel = $(document.activeElement).closest('.choiceitem');
        itemel.data('view').remove();
        this.recomputeModel();
        return true;
    },
    setupDOM: function() {
        this.$list = $(document.createElement('div'))
        this.$el.append(this.$list);
        this.$el.append(
            $(document.createElement('button'))
                .addClass('btn')
                .addClass('icon-plus')
                .addClass('append-item')
                .attr('type', 'button')
        );
    },
});
