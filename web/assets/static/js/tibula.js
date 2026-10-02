// Copyright (C) 2007-2026 by Ubaldo Porcheddu <ubaldo@eja.it>


function tableRowCheck(obj) {
  input = obj.querySelector('input')
  input.checked = !input.checked
}

function tableRowsCheck(obj) {
  document.querySelectorAll('td').forEach(function(element) {
    elementInput = element.querySelector('input')
    if (elementInput) {
      elementInput.checked = obj.checked
    }
  })
}

function tableInputCheck(obj) {
  obj.checked = !obj.checked
}

function tableEdit(obj) {
  document.querySelectorAll('td').forEach(function(element) {
    elementInput = element.querySelector('input')
    if (elementInput) {
      elementInput.checked = false
    }
  })
  input = obj.querySelector('input')
  input.checked = true
  document.getElementsByName('ejaAction').forEach(function(element) {
    if (element.value == 'edit') {
      element.click()
    }
  })
}

function fieldUpload(name) {
  var el = window._protected_reference = document.createElement('INPUT');
  el.type = 'file';
  el.addEventListener('change', function(ev) {
    var input=ev.target;
    var reader = new FileReader();
    reader.onload = function() {
    document.forms[0].elements['ejaValues['+name+']'].value=reader.result
    };
    reader.readAsText(input.files[0]);
  });

  el.click();
}

function fieldDownload(o, name) {
  var fileName=prompt('Save As');
  if (fileName) {
    o.setAttribute('href', 'data:text/plain;charset=utf-8,' + encodeURIComponent( document.forms[0].elements['ejaValues['+name+']'].value ));
    o.setAttribute('download', fileName);
    return true;
  } else {
    return false;
  }
}

function fieldEditor(name) {
  var o = document.getElementsByName('ejaValues['+name+']')[0]
  if (! editors.hasOwnProperty(name)) {
    editors[name] = SUNEDITOR.create(o, {
      buttonList: [
        ['fullScreen','undo', 'redo'],
        ['font', 'fontSize', 'formatBlock'],
        ['paragraphStyle', 'blockquote'],
        ['bold', 'underline', 'italic', 'strike', 'subscript', 'superscript'],
        ['fontColor', 'hiliteColor', 'textStyle'],
        ['removeFormat'],
        ['outdent', 'indent'],
        ['align', 'horizontalRule', 'list', 'lineHeight'],
        ['table', 'link', 'image', 'video', 'audio'],
        ['showBlocks', 'codeView', 'print'],
      ]
    })
  }
}

function fieldCalendar(name) {
  const iframeHtml = `
    <iframe 
      src="/static/calendar.html?id=${name}"
      class="w-100 h-100 border-0"
      style="min-height:100vh">
    </iframe>
  `;
  
  const closeBtn = `
    <button 
      class="btn-close position-fixed top-0 end-0 m-3"
      onclick="this.closest('.modal').remove()">
    </button>
  `;
  
  const modalHtml = `
    <div class="modal fade show d-block">
      <div class="modal-dialog modal-fullscreen">
        <div class="modal-content bg-white">
          ${iframeHtml}
          ${closeBtn}
        </div>
      </div>
    </div>
  `;
  
  document.body.insertAdjacentHTML('beforeend', modalHtml);
}

function initInactivityLogout(idle = 30, warn = 5) {
  document.body.insertAdjacentHTML('beforeend', `
    <div class="modal fade" id="idleModal" data-bs-backdrop="static">
      <div class="modal-dialog modal-dialog-centered">
        <div class="modal-content p-4 text-center">
          <h5>Session Expiring</h5>
          <p class="mb-3">Logging out in ${warn} min due to inactivity.</p>
          <button class="btn btn-primary" data-bs-dismiss="modal">Stay Logged In</button>
        </div>
      </div>
    </div>`);

  const el = document.getElementById('idleModal');
  const modal = new bootstrap.Modal(el);
  let tw, tl;

  const reset = () => {
    clearTimeout(tw);
    clearTimeout(tl);
    modal.hide();
    tw = setTimeout(() => {
      modal.show();
      tl = setTimeout(() => {
        window.onbeforeunload = null;
        window.location.href = window.location.origin + window.location.pathname;
      }, warn * 6e4);
    }, (idle - warn) * 6e4);
  };

  el.addEventListener('hidden.bs.modal', reset);
  ['mousemove', 'keydown', 'scroll', 'touchstart'].forEach(e => {
    window.addEventListener(e, () => !el.classList.contains('show') && reset(), { passive: true });
  });

  reset();
}

function formInit() {
  const f = document.getElementById('ejaForm');
  const o = {};
  
  Array.from(f.elements).forEach(function(i) {
    o[i.name] = i.value;
  });
  
  f.oninput = function() {
    window.onbeforeunload = function() {
      return 'Unsaved changes';
    };
  };

  f.onsubmit = function() {
    window.onbeforeunload = null;
  };
  
  window.onbeforeunload = function() {
    return null;
  };

  initInactivityLogout(30, 5);
}


var editors = [];


document.querySelectorAll('.toast').forEach(toast => {
  var toastInstance = new bootstrap.Toast(toast);
  setTimeout(function () {
    toastInstance.hide();
  }, 5000);
});

document.querySelectorAll('select[multiple]').forEach(select => {
  new SlimSelect({
    select: select,
    settings: {
      placeholderText: '',
    }
  });
});

document.getElementById('ejaForm')?.addEventListener('submit', function() {
  this.querySelectorAll('select').forEach(select => {
    if (select.selectedIndex === -1 || select.value === '') {
      select.value = '';
    }
  });
  for (var key in editors) {
    editors[key].save();
  }
});

window.onload = function() {
  if (document.getElementsByName('ejaGoogleSsoId').length > 0) {
  google.accounts.id.initialize({
    client_id: document.getElementsByName('ejaGoogleSsoId')[0].value,
    callback: function(e) {
      document.getElementsByName('ejaValues[googleSsoToken]')[0].value=e.credential
      document.getElementById('ejaForm').submit()
    }
  })
  google.accounts.id.renderButton(document.getElementById("google"), {type: "icon"})
  }
  formInit();
}
