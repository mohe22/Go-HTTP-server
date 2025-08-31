const loginBtn = document.getElementById("loginBtn");

loginBtn.addEventListener("click", async () => {
  const username = document.getElementById("username").value.trim();
  const password = document.getElementById("password").value.trim();
  const errorMsg = document.getElementById("errorMsg");

  if (username === "" || password === "") {
    errorMsg.textContent = "Please fill in all fields";
    errorMsg.style.display = "block";
    return;
  }

  try {
    const res = await fetch("/login", {
      method: "POST",
      headers: { "Content-Type": "application/x-www-form-urlencoded" },
      body: `username=${encodeURIComponent(username)}&password=${encodeURIComponent(password)}`,
    });

    const data = await res.json();

    if (res.ok) {
      alert(data.message);
      errorMsg.style.display = "none";
    } else {
      errorMsg.textContent = data.message;
      errorMsg.style.display = "block";
    }
  } catch (err) {
    errorMsg.textContent = "Server error";
    errorMsg.style.display = "block";
  }
});
