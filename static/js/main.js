class SecretManager {
  constructor() {
    this.initializeValidForOptions();
    this.initializeFileInput();
    this.initializeBlurToggle();
  }


  /**
   * Initializes the "valid for" options by fetching them from the API.
   */
  initializeValidForOptions() {
    // Get expiration options from API
    doCall("GET", "/api/valid-for-options", null, (status, response) => {
      if (status !== 200) {
        console.error("Failed to fetch valid-for options. Status:", status, "Response:", response);
        return;
      }
      const JsonResponse = JSON.parse(response);

      let validForElement = document.getElementById("valid-for");

      for (const key in JsonResponse) {
        let opt = document.createElement('option');
        opt.value = key;
        opt.innerHTML = `Valid for ${JsonResponse[key]}`;
        validForElement.appendChild(opt);
      }
    });
  }

  /**
   * Initializes the blur toggle functionality for the secret content textarea.
   */
  initializeBlurToggle() {
    const localStorageKey = "keepBlurredOnFocus";
    const classKeepBlurred = "keep-blurred-on-focus";

    if (this.secretContent && this.keepBlurredToggle) {
      const shouldKeepBlurred = localStorage.getItem(localStorageKey) === "true";

      this.keepBlurredToggle.checked = shouldKeepBlurred;
      this.secretContent.classList.toggle(classKeepBlurred, shouldKeepBlurred);

      this.keepBlurredToggle.addEventListener("change", () => {
        const isChecked = this.keepBlurredToggle.checked;

        this.secretContent.classList.toggle(classKeepBlurred, isChecked);
        localStorage.setItem(localStorageKey, isChecked);

        // If textarea is currently focused and toggle is unchecked, unblur it
        if (!isChecked && document.activeElement === this.secretContent) {
          this.secretContent.classList.remove("blurred");
        }
      });

      this.secretContent.addEventListener("focus", () => {
        if (!this.keepBlurredToggle.checked) {
          this.secretContent.classList.remove("blurred");
        }
      });

      this.secretContent.addEventListener("blur", () => {
        // Always add blurred class on blur, the focus listener will remove it if needed.
        this.secretContent.classList.add("blurred");
      });
    }
  }

  /**
   * Initializes the file input handling.
   */
  initializeFileInput() {
    // Handle file input changes
    const fileInput = document.getElementById("files");
    const fileList = document.getElementById("file-list");
    const fileInputContainer = document.querySelector(".file-input-container");

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
            const fileNameDiv = document.createElement("div");
            fileNameDiv.className = "file-name";
            fileNameDiv.textContent = file.name;

            const fileSizeDiv = document.createElement("div");
            fileSizeDiv.className = "file-size";
            fileSizeDiv.textContent = fileSize;

            const fileContainerDiv = document.createElement("div");
            fileContainerDiv.appendChild(fileNameDiv);
            fileContainerDiv.appendChild(fileSizeDiv);

            li.appendChild(fileContainerDiv);

            removeButton.setAttribute("aria-label", `Remove ${file.name}`);
            removeButton.addEventListener("click", (event) => {
              event.preventDefault();
              this.removeFile(file.name);
            });
            li.appendChild(removeButton);

            fileList.appendChild(li);
          }
        }
      });
    }

    if (fileInputContainer) {
      fileInputContainer.addEventListener("dragover", (e) => {
        e.preventDefault();
        fileInputContainer.classList.add("drag-over");
      }, false);

      fileInputContainer.addEventListener("dragleave", () => {
        fileInputContainer.classList.remove("drag-over");
      }, false);

      fileInputContainer.addEventListener("drop", (e) => {
        e.preventDefault();
        fileInputContainer.classList.remove("drag-over");

        const dataTransfer = new DataTransfer();
        const existingFileNames = new Set();

        for (const file of fileInput.files) {
          dataTransfer.items.add(file);
          existingFileNames.add(file.name);
        }

        for (const file of e.dataTransfer.files) {
          if (!existingFileNames.has(file.name)) {
            dataTransfer.items.add(file);
            existingFileNames.add(file.name);
          }
        }

        fileInput.files = dataTransfer.files;

        fileInput.dispatchEvent(new Event('change', { 'bubbles': true }));
      }, false);
    }
  }

  /**
   * Formats file size from bytes to a human-readable string.
   * @param {number} bytes - The file size in bytes.
   * @returns {string} The formatted file size.
   */
  formatFileSize(bytes) {
    if (bytes === 0) return '0 Bytes';

    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));

    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  }

  /**
   * Removes a file from the file input list by its name.
   * @param {string} name - The name of the file to remove.
   */
  removeFile(name) {
    const fileInput = document.getElementById("files");
    const files = fileInput.files;
    const dt = new DataTransfer();

    for (let i = 0; i < files.length; i++) {
      if (files[i].name !== name) {
        dt.items.add(files[i]);
      }
    }

    fileInput.files = dt.files;
    fileInput.dispatchEvent(new Event('change', { 'bubbles': true }));
  }

  /**
   * @returns {HTMLElement | null} The main container element.
   */
  get mainContainer() {
    return document.getElementById("main");
  }

  /**
   * @returns {HTMLElement | null} The secret content textarea element.
   */
  get secretContent() {
    return document.getElementById("secret-content");
  }

  /**
   * @returns {HTMLInputElement | null} The keep blurred toggle checkbox element.
   */
  get keepBlurredToggle() {
    return document.getElementById("keep-blurred-toggle");
  }

  /**
   * @returns {HTMLElement | null} The read secret container element.
   */
  get readSecretContainer() {
    return document.getElementById("read-secret");
  }

  /**
   * @returns {HTMLTextAreaElement | null} The read secret content textarea element.
   */
  get readSecretContent() {
    return document.getElementById("read-secret-content");
  }

  /**
   * @returns {HTMLElement | null} The error display element.
   */
  get error() {
    return document.getElementById("error");
  }

  /**
   * @returns {HTMLElement | null} The create secret container element.
   */
  get createSecretContainer() {
    return document.getElementById("create-secret");
  }

  /**
   * @returns {HTMLElement | null} The reveal secret text element.
   */
  get revealSecretText() {
    return document.getElementById("revealSecret");
  }

  /**
   * @returns {HTMLElement | null} The download files element.
   */
  get downloadFiles() {
    return document.getElementById("download-files");
  }

  /**
   * @returns {HTMLElement | null} The download files container element.
   */
  get downloadFilesContainer() {
    return document.getElementById("download-files-container");
  }

  /**
   * Displays the new secret creation view.
   */
  displayNewSecret() {
    this.hideAll();
    this.createSecretContainer.style.display = "block";
  }

  /**
   * Displays a secret by its ID.
   * @param {string} id - The ID of the secret to display.
   */
  displaySecret(id) {
    this.hideAll();
    this.readSecretContent.value = `Your secret will appear here`;
    this.readSecretContainer.style.display = "block";

    this.revealSecretText.addEventListener("click", () => {
      // Add loading state
      this.revealSecretText.innerHTML = `<div class="secret-reveal-text"><span class="secret-reveal-icon">${createFeatherIcon("loader")}</span><span>Loading...</span></div>`;

      doCall("GET", "/api/" + id, null, (status, response) => {
        if (status === 410) {
          const errorTitle = "Secret Not Found";
          const errorMessage = "This secret is no longer available. It has either already been read or has expired.";
          this.displayError(errorTitle, errorMessage);
          return;
        } else if (status !== 200) {
          const errorTitle = "An error occurred when trying to fetch the secret";
          const errorMessage = response;
          this.displayError(errorTitle, errorMessage);
          return;
        }

        response = JSON.parse(response);
        this.readSecretContent.value = response.content;

        if (response.files != null && Object.keys(response.files).length > 0) {
          this.downloadFiles.innerHTML = ''; // Clear previous content

          for (const [key, value] of Object.entries(response.files)) {
            const downloadItem = document.createElement("a");
            downloadItem.className = "download-item";
            downloadItem.href = "data:text/plain;base64," + value;
            downloadItem.download = key;
            downloadItem.innerHTML = `
              <span class="download-icon">${createFeatherIcon("download")}</span>
              <span>${key}</span>
            `;

            this.downloadFiles.appendChild(downloadItem);
          }

          this.downloadFilesContainer.style.display = "block";
        }

        this.revealSecretText.style.display = "none";
        this.readSecretContent.classList.add('active');
      });
    }, false);
  }

  /**
   * @param {string} title - Title to display when an error happens.
   * @param {string} message - Message to display when an error happens.
   */
  displayError(title, message) {
    this.error.querySelector(".error-title").textContent = title;
    this.error.querySelector(".error-message").textContent = message;

    this.hideAll();
    this.error.style.display = "block";
  }

  /**
   * Hides all main content divs.
   */
  hideAll() {
    const divs = this.mainContainer.querySelectorAll("#create-secret, #read-secret, #error");
    divs.forEach(div => {
      div.style.display = "none";
    });
  }

  /**
   * Creates a new secret.
   * @param {string} content - The content of the secret.
   * @param {number} expiresIn - The expiration time in seconds.
   */
  createSecret(content, expiresIn) {
    const data = JSON.stringify({
      "content": content,
      "expires_in": expiresIn
    });

    doCall("POST", "/api", data, (status, response) => {
      if (status < 200 || status >= 300) {
        const errorTitle = "An error occurred when trying to create the secret";
        const errorMessage = response;
        this.displayError(errorTitle, errorMessage);

        return;
      }

      let urlBox = document.getElementById("share-url");
      let shareUrl = window.location.origin + "/?id=" + response;
      urlBox.value = shareUrl;

      // Reset button state
      const saveButton = document.getElementById("save");
      saveButton.innerHTML = `<span class="button-icon">${createFeatherIcon("check")}</span><span>Success!</span>`;

      // Show the share dialog
      document.body.className += ' active';
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


/**
 * Updates the save button's disabled state based on the input element's value.
 * @param {HTMLInputElement} el - The input element to check.
 */
function updateSaveButton(el) {
  document.getElementById("save").disabled = el.value.trim() === "";
}

/**
 * Generates an SVG element with the specified icon.
 * @param {string} icon - The name of the icon to use from feather-sprite.svg.
 * @returns {string} The svg element, as a string.
 */
function createFeatherIcon(icon) {
  return `<svg class="feather"><use href="feather-sprite.svg#${icon}" /></svg>`;
}

/**
 * Makes an XMLHttpRequest.
 * @param {string} type - The HTTP method (e.g., "GET", "POST").
 * @param {string} url - The URL to request.
 * @param {string | FormData | null} data - The data to send with the request.
 * @param {function(number, string): void} fn - The callback function to execute when the request is done.
 */
function doCall(type, url, data, fn) {
  const xhr = new XMLHttpRequest();
  xhr.withCredentials = true;

  xhr.addEventListener("readystatechange", function () {
    if (this.readyState === this.DONE) {
      fn(this.status, this.responseText);
    }
  });

  xhr.open(type, url);
  xhr.setRequestHeader("Cache-Control", "no-cache, no-store, max-age=0");

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

try {
  // Handle URL box interactions
  let urlBox = document.getElementById("share-url");
  let copySuccess = document.getElementById("copy-success");

  /**
   * Shows a success message when text is copied to the clipboard.
   */
  function showCopySuccess() {
    copySuccess.classList.add("visible");
    setTimeout(() => {
      copySuccess.classList.remove("visible");
    }, 2000);
  }

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

  // Create secret button
  document.getElementById("save").addEventListener("click", function () {
    // Change button state to loading
    const saveButton = document.getElementById("save");
    const originalContent = saveButton.innerHTML;
    saveButton.disabled = true;
    saveButton.innerHTML = `<span class="button-icon">${createFeatherIcon("loader")}</span><span>Creating...</span>`;

    let content = document.getElementById("secret-content").value;
    let expires_in = parseInt(document.getElementById("valid-for").value);

    window.secretManager.createSecret(content, expires_in);

    // Reset button after timeout (in case of error)
    setTimeout(() => {
      if (saveButton.disabled) {
        saveButton.innerHTML = originalContent;
        saveButton.disabled = false;
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
