class SecretManager {
  constructor() {
    this.initializeFeatherIcons();
    this.initializeValidForOptions();
    this.initializeFileInput();
  }

  initializeFeatherIcons() {
    // Initialize feather icons
    if (typeof feather !== 'undefined') {
      feather.replace();
    }
  }

  initializeValidForOptions() {
    // Get expiration options from API
    doCall("GET", "/api/valid-for-options", null, (status, response) => {
      const JsonResponse = JSON.parse(response);

      let validForElement = document.getElementById("valid-for");

      for (const key in JsonResponse) {
        var opt = document.createElement('option');
        opt.value = key;
        opt.innerHTML = `Valid for ${JsonResponse[key]}`;
        validForElement.appendChild(opt);
      }
    });
  }

  initializeFileInput() {
    // Handle file input changes
    const fileInput = document.getElementById("files");
    const fileList = document.getElementById("file-list");

    if (fileInput) {
      fileInput.addEventListener("change", (e) => {
        fileList.innerHTML = ""; // Clear the list
        const files = e.target.files;

        if (files.length > 0) {
          for (let i = 0; i < files.length; i++) {
            const file = files[i];
            const fileSize = this.formatFileSize(file.size);

            const li = document.createElement("li");
            li.className = "file-item";
            li.innerHTML = `
              <div>
                <div class="file-name">${file.name}</div>
                <div class="file-size">${fileSize}</div>
              </div>
            `;

            fileList.appendChild(li);
          }
        }
      });
    }
  }

  formatFileSize(bytes) {
    if (bytes === 0) return '0 Bytes';

    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));

    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  }

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

  get downloadFiles() {
    return document.getElementById("download-files");
  }

  get downloadFilesContainer() {
    return document.getElementById("download-files-container");
  }

  displayNewSecret() {
    this.hideAll();
    this.createSecretContainer.style.display = "block";
  }

  displaySecret(id) {
    this.hideAll();
    this.readSecretContent.value = `Your secret will appear here`;
    this.readSecretContainer.style.display = "block";

    const that = this;
    this.revealSecretText.addEventListener("click", function () {
      // Add loading state
      that.revealSecretText.innerHTML = '<div class="secret-reveal-text"><span class="secret-reveal-icon"><i data-feather="loader"></i></span><span>Loading...</span></div>';
      feather.replace();

      doCall("GET", "/api/" + id, null, function (status, response) {
        if (status === 410) {
          that.displayNotFound();
          return;
        }

        response = JSON.parse(response);
        that.readSecretContent.value = response.content;

        if (response.files != null && Object.keys(response.files).length > 0) {
          that.downloadFiles.innerHTML = ''; // Clear previous content

          for (const [key, value] of Object.entries(response.files)) {
            const downloadItem = document.createElement("a");
            downloadItem.className = "download-item";
            downloadItem.href = "data:text/plain;base64," + value;
            downloadItem.download = key;
            downloadItem.innerHTML = `
              <span class="download-icon"><i data-feather="download"></i></span>
              <span>${key}</span>
            `;

            that.downloadFiles.appendChild(downloadItem);
          }

          that.downloadFilesContainer.style.display = "block";
          feather.replace();
        }

        that.revealSecretText.style.display = "none";
        that.readSecretContent.classList.add('active');
      });
    }, false);
  }

  displayNotFound() {
    this.hideAll();
    this.notFound.style.display = "block";
  }

  hideAll() {
    const divs = this.mainContainer.querySelectorAll("#create-secret, #read-secret, #not-found");
    divs.forEach(div => {
      div.style.display = "none";
    });
  }
}
window.secretManager = new SecretManager();


const params = new Proxy(new URLSearchParams(window.location.search), {
  get: (searchParams, prop) => searchParams.get(prop),
});

if (params.id) {
  window.secretManager.displaySecret(params.id);
} else {
  window.secretManager.displayNewSecret();
}






var updateSaveButton = function (el) {
  document.getElementById("save").disabled = el.value.trim() === "";
};

try {
  // Handle URL box interactions
  var urlBox = document.getElementById("share-url");
  var copySuccess = document.getElementById("copy-success");

  urlBox.addEventListener("click", function () {
    urlBox.focus();
    urlBox.select();
  }, false);

  // Copy button functionality
  document.getElementById("copy-button").addEventListener("click", function () {
    urlBox.focus();
    urlBox.select();

    try {
      // Modern clipboard API
      navigator.clipboard.writeText(urlBox.value).then(function () {
        showCopySuccess();
      });
    } catch (err) {
      // Fallback for older browsers
      document.execCommand("copy");
      showCopySuccess();
    }
  });

  function showCopySuccess() {
    copySuccess.classList.add("visible");
    setTimeout(() => {
      copySuccess.classList.remove("visible");
    }, 2000);
  }

  // Create secret button
  document.getElementById("save").addEventListener("click", function () {
    // Change button state to loading
    const saveButton = document.getElementById("save");
    const originalContent = saveButton.innerHTML;
    saveButton.disabled = true;
    saveButton.innerHTML = '<span class="button-icon"><i data-feather="loader"></i></span><span>Creating...</span>';
    feather.replace();

    var content = document.getElementById("secret-content").value;
    var expires_in = parseInt(document.getElementById("valid-for").value);

    createSecret(content, expires_in);

    // Reset button after timeout (in case of error)
    setTimeout(() => {
      if (saveButton.disabled) {
        saveButton.innerHTML = originalContent;
        saveButton.disabled = false;
        feather.replace();
      }
    }, 10000);
  }, false);

  // Close modal when clicking overlay
  document.getElementById("overlay").addEventListener("click", function () {
    document.body.className = '';
  }, false);
} catch (error) {
  console.error("Error setting up event handlers:", error);
}

function createSecret(content, expiresIn) {
  const data = JSON.stringify({
    "content": content,
    "expires_in": expiresIn
  });

  doCall("POST", "/api", data, function (status, response) {
    var shareUrl = window.location.origin + "/?id=" + response;
    urlBox.value = shareUrl;

    // Reset button state
    const saveButton = document.getElementById("save");
    saveButton.innerHTML = '<span class="button-icon"><i data-feather="check"></i></span><span>Success!</span>';
    feather.replace();

    // Show the share dialog
    document.body.className += ' active';
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

  if (data) {
    const files = document.getElementById("files").files;
    const formData = new FormData();
    formData.append("data", data);
    for (const file of files) {
      formData.append("files", file);
    }

    xhr.send(formData);
  } else {
    xhr.send();
  }
}

// Initialize feather icons after page load
document.addEventListener('DOMContentLoaded', function () {
  if (typeof feather !== 'undefined') {
    feather.replace();
  }
});
