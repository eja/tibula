// Copyright (C) 2007-2025 by Ubaldo Porcheddu <ubaldo@eja.it>


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
  setTimeout(()=>{ alert("Logging out for inactivity in a minute"); }, 950*1000);
  setTimeout(()=>{ window.location.href = window.location.origin + window.location.pathname; }, 1000*1000);
  
  formInit();
}
