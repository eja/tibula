// Copyright (C) 2007-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

function tableRowCheck(obj) {
	input = obj.querySelector("input")
	input.checked = !input.checked
}

function tableRowsCheck(obj) {
	document.querySelectorAll("td").forEach(function(element) {
		elementInput = element.querySelector("input")
		if (elementInput) {
			elementInput.checked = obj.checked
		}
	})
}

function tableInputCheck(obj) {
	obj.checked = !obj.checked
}

function tableEdit(obj) {
  document.querySelectorAll("td").forEach(function(element) {
    elementInput = element.querySelector("input")
		if (elementInput) {
    	elementInput.checked = false
		}
  })
	input = obj.querySelector("input")
  input.checked = true
	document.getElementsByName('ejaAction').forEach(function(element) {
		if (element.value == "edit") {
			element.click()
		}
	})
}

function fieldUpload(name) {
 var el = window._protected_reference = document.createElement("INPUT");
 el.type = "file";
 el.addEventListener('change', function(ev) {
  var input=ev.target;
  var reader = new FileReader();
  reader.onload = function() {
   document.forms[0].elements["ejaValues["+name+"]"].value=reader.result
  };
  reader.readAsText(input.files[0]);
 });

 el.click();
}

function fieldDownload(o, name) {
 var fileName=prompt("Save As");
 if (fileName) {
  o.setAttribute('href', 'data:text/plain;charset=utf-8,' + encodeURIComponent( document.forms[0].elements["ejaValues["+name+"]"].value ));
  o.setAttribute('download', fileName);
  return true;
 } else {
  return false;
 }
}

var toasts = document.querySelectorAll('.toast');
toasts.forEach(function (toast) {
	var toastInstance = new bootstrap.Toast(toast);
	setTimeout(function () {
		toastInstance.hide();
	}, 5000);
});
