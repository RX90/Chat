document.getElementById('signup-form').addEventListener('submit', async function (e) {
  e.preventDefault();

  const email = document.getElementById('email').value;
  const username = document.getElementById('username').value;
  const password = document.getElementById('password').value;

  try {
    const response = await fetch('/api/auth/sign-up', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ email, username, password })
    });

    if (response.ok) {
      alert('Registered successfully!');
      this.reset();
    } else {
      alert('Registration failed with status: ' + response.status);
    }
  } catch (err) {
    alert('Error: ' + err.message);
  }
});
