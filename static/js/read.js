var secretContent = document.getElementById("secret-content");
var notFound = document.getElementById("not-found");

const xhr = new XMLHttpRequest();
xhr.withCredentials = true;

xhr.addEventListener("readystatechange", function () {
  if (this.readyState === this.DONE) {
    if (this.status === 410) {
      secretContent.style.display = "none";
      notFound.style.display = "block";
      return;
    }
    secretContent.value = this.responseText
    secretContent.style.display = "inline-block";
  }
});

xhr.open("GET", window.location.pathname.split("/")[1]);
xhr.setRequestHeader("Content-Type", "application/json");

xhr.send();