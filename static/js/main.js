var save = function() {
  const data = JSON.stringify({
    "content": document.getElementById("secret-content").value,
    "expires_in": parseInt(document.getElementById("valid-for").value)
  });

  const xhr = new XMLHttpRequest();
  xhr.withCredentials = true;

  xhr.addEventListener("readystatechange", function () {
    if (this.readyState === this.DONE) {
      var shareUrl = window.location.origin + "/" + this.responseText;
      var urlBox = document.getElementById("share-url");
      urlBox.value = shareUrl;
      document.getElementById("result-container").style.display = "flex";
      
      copyShareUrl()
    }
  });

  xhr.open("POST", "http://localhost:8080/");
  xhr.setRequestHeader("Content-Type", "application/json");

  xhr.send(data);
};


var clip = function(el) {
  var range = document.createRange();
  range.selectNodeContents(el);
  var sel = window.getSelection();
  sel.removeAllRanges();
  sel.addRange(range);
};

var addToClipboard = function(text) {
  try {
    navigator.clipboard.writeText(text);
  } catch (err) {
    console.error('Failed to copy: ', err);
  }
};

var copyShareUrl = function() {
  var urlBox = document.getElementById("share-url");
  urlBox.select();
  var shareUrl = urlBox.value;
  console.log(shareUrl)
  addToClipboard(shareUrl)
};

