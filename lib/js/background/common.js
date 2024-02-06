var core = {
  "start": function () {
    core.load();
  },
  "install": function () {
    core.load();
  },
  "action": {
    "contextmenu": function (e) {
      if (e.menuItemId === "test.page") {
        app.tab.open(config.test.page);
      } else {
        config.notification.show = !config.notification.show;
      }
    },
    "storage": function (changes, namespace) {
      if ("notification" in changes) {
        app.contextmenu.update("notification.checkbox", {
          "checked": config.notification.show,
        }, app.error);
      }
    }
  },
  "load": function () {
    app.contextmenu.create({
      "type": "normal",
      "id": "test.page",
      "contexts": ["browser_action"],
      "title": "What is my Fingerprint?"
    }, app.error);
    /*  */
    app.contextmenu.create({
      "type": "checkbox",
      "id": "notification.checkbox",
      "contexts": ["browser_action"],
      "checked": config.notification.show,
      "title": "Show Desktop Notifications"
    }, app.error);
  }
};

app.popup.receive("support", function () {app.tab.open(app.homepage())});
app.popup.receive("fingerprint", function () {app.tab.open(config.test.page)});
app.popup.receive("donation", function () {app.tab.open(app.homepage() + "?reason=support")});

app.popup.receive("load", function () {
  app.popup.send("storage", {
    "notifications": config.notification.show
  });
});

app.popup.receive("notifications", function () {
  config.notification.show = !config.notification.show;
  app.popup.send("storage", {
    "notifications": config.notification.show
  });
});

app.page.receive("fingerprint", function (e) {
  const message = "\nA fingerprinting attempt is detected!\nYour browser is reporting a fake value.";
  //
  if (config.notification.show) {
    if (config.notification.timeout) clearTimeout(config.notification.timeout);
    config.notification.timeout = setTimeout(function () {
      app.notifications.create({
        "type": "basic",
        "title": app.name(),
        "message": e.host + message
      });
    }, 1000);
  }
});

app.on.startup(core.start);
app.on.installed(core.install);
app.on.storage(core.action.storage);
app.contextmenu.on.clicked(core.action.contextmenu);