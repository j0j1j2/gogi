package webclient

const Script = `(function(){
async function request(path, options) {
  const response = await fetch(path, options || {});
  const text = await response.text();
  let body = null;
  if (text) {
    try {
      body = JSON.parse(text);
    } catch (_) {
      body = text;
    }
  }
  if (!response.ok) {
    const message = body && body.error ? body.error : (typeof body === "string" ? body : response.statusText);
    throw new Error(message);
  }
  return body;
}

window.gogi = {
  state: function() {
    return request("/api/state");
  },
  toggle: function(id, enabled) {
    return request("/api/toggle/" + encodeURIComponent(id) + "?enabled=" + encodeURIComponent(String(enabled)), {
      method: "POST"
    });
  },
  action: function(id, payload) {
    return request("/api/action/" + encodeURIComponent(id), {
      method: "POST",
      headers: {"content-type": "application/json"},
      body: JSON.stringify(payload == null ? {} : payload)
    });
  }
};
})();`
