Ext.require([
    'Ext.Viewport',
    'Ext.data.JsonP',
    'Ext.data.JsonStore',
    'Ext.tip.QuickTipManager',
    'Ext.tab.Panel'
]);



Ext.define('My.Notifyclient', {
    mixins: {
        observable: 'Ext.util.Observable'
    },

    constructor: function(config) {
        this.mixins.observable.constructor.call(this, config);
        this.addEvents('task');
        this.boxid = null;
        this.url = config.url;
        this.failures = 0;
        this.req();
    },

    req: function() {
        Ext.Ajax.request({
            url: this.url,
            method: 'GET',
            params : {
                boxid: this.boxid,
                immediate: this.failures > 0 ? 1 : 0
            },
            success: function(result, request) {
                var response = Ext.JSON.decode(result.responseText);
                this.boxid = response.boxid;
                for (var x = 0; x < response.messages.length; x++) {
                    var m = response.messages[x];
                    console.log(m.type, m.op, m.obj);
                    this.fireEvent(m.type, m.op, m.obj);
                }
                if(this.failures > 1) {
                    console.log('reconnected');
                }
                this.failures = 0;
                this.delay(100);
            }.bind(this),
            failure: function(result, request) {
                this.delay(this.failures ? 5000 : 1);
                if(this.failures == 1) {
                    console.log('disconnected');
                }
                this.failures++;
            }.bind(this)
        });
    },

    delay: function(time) {
        var task = new Ext.util.DelayedTask(function() {
            this.req();
        }.bind(this));
        task.delay(time);
    }
});



Ext.define('My.Plugin', {
    extend: 'Ext.data.Model',
    fields: ['id', 'owner', 'upstream', 'betasecret', 'downloadurl']
});


Ext.define('My.PluginVersion', {
    idProperty: 'version',
    extend: 'Ext.data.Model',
    fields: ['version', 'title', 'type', 'author', 'showtimeVersion', 
	     'published', 'status', 'downloads']
});



var pluginstore = Ext.create('Ext.data.Store', {
    autoLoad: true,
    model: 'My.Plugin',
    sorters: [{
        property: 'id',
        direction: 'ASC'
    }],
    proxy: {
        type: 'ajax',
        url: 'plugins',
        reader: {
            type: 'json',
            root: 'plugins'
        }
    }
});


var plugineditor = function(rec, user) {

    var do_on_selection = function(op, reason) {
	var vrec = versionlist.getSelectionModel().getSelection()[0].getData();
	Ext.Ajax.request({
	    url: 'versions/' + rec.id + '/' + vrec.version,
	    params: {
		op: op,
		reason: reason
	    },
	    success: function() {
		store.reload();
	    }
	})	
    }

    var del = Ext.create('Ext.Action', {
	icon: 'static/icons/delete.png',
        text: 'Delete',
        disabled: true,
        handler: function(widget, event) {
	    var ver = versionlist.getSelectionModel().getSelection()[0].getData().version;
	    Ext.MessageBox.confirm('Delete version "' + ver +'"', 'Are you sure', function(res) {
		if(res == 'yes') 
		    do_on_selection('delete');
	    });
        }
    });

    var approve = Ext.create('Ext.Action', {
	icon: 'static/icons/accept.png',
        text: 'Approve',
        disabled: true,
        handler: function(widget, event) {
	    do_on_selection('approve');
        }
    });


    var unapprove = Ext.create('Ext.Action', {
	icon: 'static/icons/clock.png',
        text: 'Set pending',
        disabled: true,
        handler: function(widget, event) {
	    do_on_selection('unapprove');
        }
    });

    var reject = Ext.create('Ext.Action', {
	icon: 'static/icons/cancel.png',
        text: 'Reject',
        disabled: true,
        handler: function(widget, event) {
	    Ext.MessageBox.show({
		title: 'Reject',
		msg: 'Enter reason:',
		width:300,
		buttons: Ext.MessageBox.OKCANCEL,
		multiline: true,
		fn: function(res, txt) {
		    if(res == 'ok')
			do_on_selection('reject', txt);
		}
	    });
        }
    });



    var publish = Ext.create('Ext.Action', {
	icon: 'static/icons/plugin_add.png',
        text: 'Publish',
        disabled: true,
        handler: function(widget, event) {
	    do_on_selection('publish');
        }
    });


    var revoke = Ext.create('Ext.Action', {
	icon: 'static/icons/plugin_delete.png',
        text: 'Revoke',
        disabled: true,
        handler: function(widget, event) {
	    do_on_selection('revoke');
        }
    });



    var actions = [del, '-', publish, revoke];
    if(user.Admin) {
	actions.push('-');
	actions.push(approve);
	actions.push(unapprove);
	actions.push(reject);
    }
    
    var contextMenu = Ext.create('Ext.menu.Menu', {
        items: actions
    });

    var store = Ext.create('Ext.data.Store', {
        autoLoad: true,
        model: 'My.PluginVersion',
	sorters: [{
            property: 'version',
            direction: 'ASC'
	}],
        proxy: {
            type: 'ajax',
            url: 'versions/' + rec.id,
            reader: {
                type: 'json',
                root: 'versions'
            }
        }
    });

    var versionlist = Ext.create('Ext.grid.Panel', {
	store: store,
        columnLines: true,
        title: 'Versions',
        region: 'center',
        columns: [{
            text     : 'Version',
            flex     : 0.5,
            dataIndex: 'version'
	}, {
            text     : 'Title',
            flex     : 1,
            dataIndex: 'title'
	}, {
            text     : 'Downloads',
            flex     : 0.5,
            dataIndex: 'downloads'
	}, {
            text     : 'Type',
            flex     : 0.5,
            dataIndex: 'type'
	}, {
            text     : 'Author',
            flex     : 1,
            dataIndex: 'author'
	}, {
            text     : 'Showtime minimum version',
            flex     : 1,
            dataIndex: 'showtimeVersion'
	}, {
	    text    : 'Status',
	    flex    : 0.5,
	    dataIndex: 'status',
	    renderer : function(value, metadata, record, row, col, store) {
		switch(value) {
		case 'a':
		    return '<span style="color:green">Approved</span>';
		case 'p':
		    return 'Pending'
		case 'r':
		    return '<span style="color:red">Rejected</span>';
		}
	    }
	}, {
	    xtype: 'booleancolumn',
	    text    : 'Published',
	    flex    : 0.5,
	    dataIndex: 'published',
	    trueText: 'Yes',
            falseText: 'No' 
	}],
	dockedItems: [{
            xtype: 'toolbar',
            items: actions
        }],
        viewConfig: {
            stripeRows: true,
            listeners: {
                itemcontextmenu: function(view, rec, node, index, e) {
                    e.stopEvent();
                    contextMenu.showAt(e.getXY());
                    return false;
                }
            }
        },
    })

    versionlist.getSelectionModel().on({
        selectionchange: function(sm, selections) {

	    function able(o, b) {
		if(b)
		    o.enable();
		else
		    o.disable();
	    }

            if (selections.length == 1) {
		d = selections[0].getData();

		able(del,        d.status == 'p');

		able(approve,    d.status != 'a');
		able(unapprove,  d.status != 'p');
		able(reject,     d.status != 'r');

		able(publish,   !d.published);
		able(revoke,     d.published);

            } else {
		for (var v in actions) 
		    if(typeof actions[v] == 'object')
			actions[v].disable();
            }
        }
    });


//---------------------------------------


    var dlbtn = Ext.create('Ext.Button', {
	text: 'Download from upstream',
	handler: function() {
	    Ext.MessageBox.show({
		msg: 'Ingest new plugin',
		progressText: 'Downloading...',
		width:300,
		wait:true,
		waitConfig: {interval:200}
	    });


	    Ext.Ajax.request({
		url: 'getFromUpstream/' + rec.id,
		success: function(response) {
		    var v = Ext.JSON.decode(response.responseText);
		    Ext.MessageBox.hide();

		    if(!v.success) {
			Ext.Msg.alert('SPMC error', v.error);
		    } else {
			console.log(v);
			autoopen = v.result.id;
			store.reload();
		    }
		},
		failure: function(fp, o) {
		    Ext.MessageBox.hide();
		    Ext.Msg.alert('Communication error', o.result.error);
		}
	    });
	}
    });

    if(!rec.downloadurl)
	dlbtn.disable();
  
    var buttons = [{
	text: 'Save',
	handler: function() {
	var form = this.up('form').getForm();
	    if(form.isValid()){
		var vals = form.getFieldValues();
		if(vals['downloadurl'])
		    dlbtn.enable();
		else
		    dlbtn.disable();
		
		form.submit({
		    url: 'plugins/' + rec.id,
	    
		    success: function(fp, o) {
			console.log('reloading');
			pluginstore.reload();
		    }
		});
	    }
	}
    }];


    buttons.push(dlbtn);

    if(user.Admin) {
	buttons.push({
	    text: 'Erase',
	    handler: function() {
		Ext.MessageBox.confirm('Erase "' + rec.id +'"', 'Are you sure', function(res) {
		    if(res == 'yes')  {

			Ext.Ajax.request({
			    method: 'POST',
			    url: 'erasePlugin/' + rec.id,			
			    success: function(response) {
				var v = Ext.JSON.decode(response.responseText);

				if(!v.success) {
				    Ext.Msg.alert('SPMC error', v.error);
				} else {
				    editor.close();
				    pluginstore.reload();
				}
			    },
			    failure: function(fp, o) {
				Ext.Msg.alert('Communication error', o.result.error);
			    }
			});
		    }
		});
	    }
	});
    }

    var form = Ext.create('Ext.form.Panel', {
	region: 'north',
        bodyPadding: 5,

	fieldDefaults: {
            labelAlign: 'right',
            labelWidth: 200,
            anchor: '100%'
        },

	items: [{
            xtype: 'textfield',
            name: 'betasecret',
            fieldLabel: 'Beta testing password',
            value: rec.betasecret
	},{
            xtype: 'textfield',
            name: 'downloadurl',
            fieldLabel: 'Upstream URL (zipfile)',
            value: rec.downloadurl
	}],


	buttons: buttons,
	buttonAlign: 'left'
    });


    var editor = Ext.create('Ext.Panel', {
	region: 'center',
	title: 'Editing plugin "' + rec.id + '"',
        layout: {
            type: 'border',
            padding: 5
        },
        defaults: {
            split: true
        },
        items: [form, versionlist]
    });
    
    return editor;
}

function readCookie(name) {
    var nameEQ = name + "=";
    var ca = document.cookie.split(';');
    for(var i=0;i < ca.length;i++) {
	var c = ca[i];
	while (c.charAt(0)==' ') c = c.substring(1,c.length);
	if (c.indexOf(nameEQ) == 0) return c.substring(nameEQ.length,c.length);
    }
    return null;
}

function decode64(input) {
    var output = "";
    var chr1, chr2, chr3 = "";
    var enc1, enc2, enc3, enc4 = "";
    var i = 0;
    var keyStr = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/=";

    input = input.replace(/[^A-Za-z0-9\+\/\=]/g, "");
 
    do {
        enc1 = keyStr.indexOf(input.charAt(i++));
        enc2 = keyStr.indexOf(input.charAt(i++));
        enc3 = keyStr.indexOf(input.charAt(i++));
        enc4 = keyStr.indexOf(input.charAt(i++));
	
        chr1 = (enc1 << 2) | (enc2 >> 4);
        chr2 = ((enc2 & 15) << 4) | (enc3 >> 2);
        chr3 = ((enc3 & 3) << 6) | enc4;
	
        output = output + String.fromCharCode(chr1);
	
        if (enc3 != 64)
            output = output + String.fromCharCode(chr2);

        if (enc4 != 64)
            output = output + String.fromCharCode(chr3);
	
        chr1 = chr2 = chr3 = "";
        enc1 = enc2 = enc3 = enc4 = "";
	
    } while (i < input.length);
    return output;
}


Ext.onReady(function () {

    var auth = readCookie("auth");
    var user = Ext.JSON.decode(decode64(auth.split('_')[0]));
    Ext.tip.QuickTipManager.init();
    var autoopen;


//    var notifier = new My.Notifyclient({url: '/comet'})
    var uploadWin;

   var uploadBtn = Ext.create('Ext.Action', {
	icon: 'static/icons/compress.png',
        text: 'Upload plugin package',
        handler: function() {
	    if(!uploadWin) {
		win = Ext.create('widget.window', {
                    title: 'Upload package',
                    closable: true,
                    closeAction: 'hide',
                    width: 600,
                    minWidth: 350,
                    height: 100,
                    layout: {
			type: 'anchor',
			padding: 5
                    },
                    items: [{
			xtype: 'form',
			bodyPadding: '10 10 0',
			border: false,
			layout: 'anchor',

			defaults: {
			    anchor: '100%',
			    allowBlank: false,
			    msgTarget: 'side',
			    labelWidth: 50
			},

			items: [{
			    xtype: 'filefield',
			    id: 'form-file',
			    emptyText: 'Select a plugin package',
			    fieldLabel: 'ZIP file',
			    name: 'plugin',
			    buttonText: '',
			    buttonConfig: {
				icon: 'static/icons/compress.png'
			    }
			}],
			
			buttons: [{
			    text: 'Upload',
			    handler: function(){
				var form = this.up('form').getForm();
				if(form.isValid()){
				    form.submit({
					url: 'upload',
					waitMsg: 'Uploading your plugin...',
					success: function(fp, o) {
					    uploadBtn.enable();
					    win.hide();
					    autoopen = o.result.result.id;
					    pluginstore.reload();
					    
					},
					failure: function(fp, o) {
					    Ext.Msg.alert('Error', o.result.error);
					}
				    });
				}
			    }
			},{
			    text: 'Reset',
			    handler: function() {
				this.up('form').getForm().reset();
			    }
			}]
                    }]
		});
            }
	    uploadBtn.disable();
            if (win.isVisible()) {
		win.hide(this, function() {
		    uploadBtn.enable();
		});
            } else {
		win.show(this, function() {
		    uploadBtn.enable();
		});
            }
        }
    });

    var pluginlist = Ext.create('Ext.grid.Panel', {
	store: pluginstore,
        title: 'Showtime plugin manager. Logged in as "' + user.Username + '"' +
	    (user.Admin ? ' [Administrator]' : ''),
        region: 'west',
        width: '30%',
        columns: [{
            text     : 'ID',
            flex     : 1,
            dataIndex: 'id'
	}, {
            text     : 'Owner',
            flex     : 1,
            dataIndex: 'owner'
	}],
	dockedItems: [{
            xtype: 'toolbar',
            items: [uploadBtn, {
		icon: 'static/icons/door_in.png',
		text: 'Logout',
		handler: function() {
		    window.open('logout','_self',false)
		}
	    }, '->', {
		icon: 'static/icons/refresh.gif',
		text: 'Refresh',
		handler: function() {
		    pluginstore.reload();
		}

		
	    }]
        }],

    })

    var current = null;

    var vp = Ext.create('Ext.Viewport', {
        layout: {
            type: 'border',
            padding: 5
        },
        defaults: {
            split: true
        },
        items: [pluginlist]
    });

    

    pluginlist.on('itemclick', function (view, rec){
	if(current)
	    vp.remove(current);

	current = vp.add(new plugineditor(rec.getData(), user));
	vp.doLayout();
    });


    pluginstore.on('load', function() {
	if(autoopen) {
	    if(current)
		vp.remove(current);

	    rec = pluginstore.getById(autoopen);
	    autoopen = null;
	    current = vp.add(new plugineditor(rec.getData(), user));
	    vp.doLayout();
	}
    });

});
