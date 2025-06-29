document.addEventListener("DOMContentLoaded", () => {
  const addBtn = document.getElementById("addBtn");
  const tableBody = document.getElementById("studentTable");
  const searchInput = document.getElementById("searchInput");

  let studentList = [];

  function renderTable(list) {
    tableBody.innerHTML = "";
    list.forEach(student => {
      const row = document.createElement("tr");
      row.innerHTML = `
        <td class="p-2">${student.application_no}</td>
        <td class="p-2">${student.candidate_name}</td>
        <td class="p-2">${student.branch}</td>
        <td class="p-2 text-green-700 font-semibold">Present</td>
      `;
      tableBody.appendChild(row);
    });
  }

  function isValidApplicationNo(appNo) {
    return /^(131|128|127)[0-9]{9}$/.test(appNo);
  }

  addBtn.addEventListener("click", async () => {
    const appNo = document.getElementById("application_no").value.trim();
    const name = document.getElementById("candidate_name").value.trim();
    const branch = document.getElementById("branch").value;

    if (!isValidApplicationNo(appNo)) {
      alert("Invalid Application Number! Must be 12 digits starting with 131, 128, or 127.");
      return;
    }

    if (!name || !branch) {
      alert("Please fill all fields.");
      return;
    }

    const newStudent = {
      application_no: appNo,
      candidate_name: name,
      branch: branch
    };

    try {
      const res = await fetch("/add_student", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(newStudent)
      });

      const result = await res.json();
      if (result.success) {
        studentList.push(newStudent);
        renderTable(studentList);
        document.getElementById("application_no").value = "";
        document.getElementById("candidate_name").value = "";
        document.getElementById("branch").value = "";
      } else {
        alert(result.message || "Failed to add student.");
      }
    } catch (err) {
      console.error(err);
      alert("Server error. Please try again.");
    }
  });

  searchInput.addEventListener("input", () => {
    const keyword = searchInput.value.trim().toLowerCase();
    const filtered = studentList.filter(s =>
      s.application_no.includes(keyword)
    );
    renderTable(filtered);
  });

  // Optional: Preload existing students (future enhancement)
  // fetch("/all_students").then(res => res.json()).then(data => {
  //   studentList = data;
  //   renderTable(studentList);
  // });
});
