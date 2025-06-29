document.addEventListener("DOMContentLoaded", () => {
  const appNoInput = document.getElementById("application_no");
  const nameInput = document.getElementById("candidate_name");
  const branchSelect = document.getElementById("branch");
  const statusSelect = document.getElementById("status");
  const addOrUpdateBtn = document.getElementById("addOrUpdateBtn");
  const saveAllBtn = document.getElementById("saveAllBtn");
  const tableBody = document.getElementById("studentTable");
  const searchInput = document.getElementById("searchInput");

  let studentList = [];
  let changedStatuses = new Map();

  function isValidApplicationNo(appNo) {
    return /^(131|128|127)[0-9]{9}$/.test(appNo);
  }

  function renderTable(data) {
    tableBody.innerHTML = "";
    data.forEach(student => {
      const row = document.createElement("tr");

      const statusOptions = `
        <select data-app="${student.application_no}" class="status-dropdown border px-2 py-1 rounded">
          <option value="Present" ${student.status === "Present" ? "selected" : ""}>Present</option>
          <option value="Reporting Slip Given" ${student.status === "Reporting Slip Given" ? "selected" : ""}>Reporting Slip Given</option>
        </select>
      `;

      row.innerHTML = `
        <td class="p-2">${student.application_no}</td>
        <td class="p-2">${student.candidate_name}</td>
        <td class="p-2">${student.branch}</td>
        <td class="p-2">${statusOptions}</td>
      `;
      tableBody.appendChild(row);
    });

    // Attach change event listeners
    document.querySelectorAll(".status-dropdown").forEach(dropdown => {
      dropdown.addEventListener("change", (e) => {
        const appNo = e.target.getAttribute("data-app");
        const newStatus = e.target.value;
        changedStatuses.set(appNo, newStatus);
      });
    });
  }

  function loadStudents() {
    fetch("/get_students")
      .then(res => res.json())
      .then(data => {
        studentList = data;
        changedStatuses.clear();
        renderTable(data);
      });
  }

  addOrUpdateBtn.addEventListener("click", async () => {
    const application_no = appNoInput.value.trim();
    const candidate_name = nameInput.value.trim();
    const branch = branchSelect.value;
    const status = statusSelect.value;

    if (!isValidApplicationNo(application_no)) {
      alert("Application number must be 12 digits starting with 131, 128, or 127.");
      return;
    }

    if (!candidate_name || !branch || !status) {
      alert("All fields are required.");
      return;
    }

    const res = await fetch("/update_student", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ application_no, candidate_name, branch, status })
    });

    const result = await res.json();
    if (result.success) {
      appNoInput.value = "";
      nameInput.value = "";
      branchSelect.value = "";
      statusSelect.value = "Present";
      loadStudents();
    } else {
      alert(result.message || "Failed to save.");
    }
  });

  saveAllBtn.addEventListener("click", async () => {
    if (changedStatuses.size === 0) {
      alert("No changes to save.");
      return;
    }

    const updates = Array.from(changedStatuses.entries()).map(([application_no, status]) => ({ application_no, status }));

    const res = await fetch("/bulk_update_status", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ updates })
    });

    const result = await res.json();
    if (result.success) {
      alert("Statuses updated successfully.");
      loadStudents();
    } else {
      alert("Update failed.");
    }
  });

  searchInput.addEventListener("input", () => {
    const keyword = searchInput.value.trim().toLowerCase();
    const filtered = studentList.filter(s => s.application_no.includes(keyword));
    renderTable(filtered);
  });

  loadStudents();
});
