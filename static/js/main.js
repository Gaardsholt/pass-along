
var updateSaveButton = function (el) {
  document.getElementById("save").disabled = el.value == "";
};


try {
  var urlBox = document.getElementById("share-url");

  urlBox.addEventListener("click", function () {
    urlBox.focus();
    urlBox.select();
  }, false);

  document.getElementById("save").addEventListener("click", function () {
    const data = JSON.stringify({
      "content": document.getElementById("secret-content").value,
      "expires_in": parseInt(document.getElementById("valid-for").value)
    });

    const xhr = new XMLHttpRequest();

    xhr.addEventListener("readystatechange", function () {
      if (this.readyState === this.DONE) {
        var shareUrl = window.location.origin + "/" + this.responseText;
        urlBox.value = shareUrl;
        urlBox.select();
        document.execCommand("copy");

        document.body.className += ' active';
      }
    });

    xhr.open("POST", "/");
    xhr.setRequestHeader("Content-Type", "application/json");

    xhr.send(data);
  }, false);

  document.getElementById("overlay").addEventListener("click", function () {
    document.body.className = '';
  }, false);
} catch (error) {

}



