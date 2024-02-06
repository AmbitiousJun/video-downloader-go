var config = {};

config.test = {"page": "https://webbrowsertools.com/webgl-fingerprint/"};

config.welcome = {
  set lastupdate (val) {app.storage.write("lastupdate", val)},
  get lastupdate () {return app.storage.read("lastupdate") !== undefined ? app.storage.read("lastupdate") : 0}
};

config.notification = {
  "timeout": null,
  set show (val) {app.storage.write("notification", val)},
  get show () {return app.storage.read("notification") !== undefined ? app.storage.read("notification") : false}
};
