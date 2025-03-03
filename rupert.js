"use strict";

function enableImportLoad(obj) {
  let ldr = document.getElementById("submitbutton");
  if (ldr) ldr.disabled = false;

  let csv = document.getElementById("jsonfile");
  if (!csv) return;
  let data = document.getElementById("json");
  if (!data) return;
  const file = csv.files[0];
  console.log("File is " + file);
  if (file) {
    const reader = new FileReader();
    reader.onload = function (e) {
      const content = e.target.result;
      data.value = content;

      let ok = RBLRCSVRE.test(content);
      console.log("ok == " + ok);
      if (!ok) {
        console.log("no match");
        return;
      }

      if (ldr) ldr.classList.remove("hide");
    };
    reader.readAsText(file);
  }
}
