class SecretManager {
  // constructor(height, width) {
  //   this.height = height;
  //   this.width = width;
  // }

  get mainContainer() {
    return document.getElementById("main");
  }
  get secretContent() {
    return document.getElementById("secret-content");
  }
  get readSecretContainer() {
    return document.getElementById("read-secret");
  }
  get readSecretContent() {
    return document.getElementById("read-secret-content");
  }
  get notFound() {
    return document.getElementById("not-found");
  }
  get createSecretContainer() {
    return document.getElementById("create-secret");
  }
  get revealSecretText() {
    return document.getElementById("revealSecret"); 
  }

  displayNewSecret() {
    this.hideAll();
    this.createSecretContainer.style = "display: block";
  }
  displaySecret(id) {
    this.hideAll();
    this.readSecretContent.value = `Your super secret
password goes here`;
    this.readSecretContainer.style = "display: block";

    const that = this;
    this.revealSecretText.addEventListener("click", function () {
      doCall("GET", "/api/" + id, null, function(status, response) {

        if (status === 410) {
          that.displayNotFound();
          return;
        }
        that.readSecretContent.value = response;
        that.revealSecretText.style = "display: none";
        that.readSecretContent.classList.add('active');
      });
  
    }, false);

  }
  displayNotFound() {
    this.hideAll();
    this.notFound.style = "display: block";
  }
  hideAll() {
    for (let cell of this.mainContainer.getElementsByTagName("div")) {
      cell.style.display = "none";
    }
  }

}
window.secretManager = new SecretManager();


const params = new Proxy(new URLSearchParams(window.location.search), {
  get: (searchParams, prop) => searchParams.get(prop),
});

if (params.id) {
  window.secretManager.displaySecret(params.id);
  // readSecret(params.id);
}else {
  window.secretManager.displayNewSecret();
}






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
    var content = document.getElementById("secret-content").value;
    var expires_in = parseInt(document.getElementById("valid-for").value)

    createSecret(content, expires_in);

  }, false);

  document.getElementById("overlay").addEventListener("click", function () {
    document.body.className = '';
  }, false);
} catch (error) {

}


function createSecret(content, expiresIn) {
  const data = JSON.stringify({
    "content": content,
    "expires_in": expiresIn
  });

  doCall("POST", "/api", data, function(status, response) {
    var shareUrl = window.location.origin + "/?id=" + response;
    urlBox.value = shareUrl;
    urlBox.select();
    document.execCommand("copy");

    document.body.className += ' active';
  });
}

function readSecret(id) {
  doCall("GET", "/api/" + id, null, function(status, response) {

    if (status === 410) {
      window.secretManager.displayNotFound();
      return;
    }
    window.secretManager.readSecretContent.value = response
    window.secretManager.displaySecret();

  });
}

function doCall(type, url, data, fn) {
  const xhr = new XMLHttpRequest();
  xhr.withCredentials = true;

  xhr.addEventListener("readystatechange", function () {
    if (this.readyState === this.DONE) {
      fn(this.status, this.responseText);
    }
  });

  xhr.open(type, url);
  xhr.setRequestHeader("Content-Type", "application/json");

  if (data) {
    xhr.send(data);
  }else {
    xhr.send();
  }
}