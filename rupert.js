"use strict";

function chooseRally(obj) {
  console.log("Choosing Rally");
  let sel = obj;
  let code = document.getElementById("rallycode");
  let desc = document.getElementById("rallydesc");

  console.log("Got the objects");
  code.value = sel.options[sel.selectedIndex].value;
  desc.value = sel.options[sel.selectedIndex].innerText;
  code.readOnly = code.value != "";
  desc.readOnly = code.readOnly;

  enableImportLoad(obj);
}

function enableImportLoad(obj) {
  let ldr = document.getElementById("submitbutton");

  let csv = document.getElementById("thefile");
  if (!csv) return;
  let data = document.getElementById("thedata");
  if (!data) return;
  const file = csv.files[0];
  console.log("File is " + file);
  if (file) {
    const reader = new FileReader();
    reader.onload = function (e) {
      const content = e.target.result;
      data.value = content;

      if (ldr) {
        ldr.disabled = false;
        ldr.classList.add("btn")
        ldr.classList.remove("hide");
      }
    };
    reader.readAsText(file);
  }
}
