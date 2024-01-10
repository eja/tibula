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

var toasts = document.querySelectorAll('.toast');
toasts.forEach(function (toast) {
	var toastInstance = new bootstrap.Toast(toast);
	setTimeout(function () {
		toastInstance.hide();
	}, 5000);
});
