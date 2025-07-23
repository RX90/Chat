document.getElementById('signin-form').addEventListener('submit', async function (e) {
  e.preventDefault();

  const email = document.getElementById('email').value;
  const password = document.getElementById('password').value;

  try {
    const response = await fetch('/api/auth/sign-in', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ email, password })
    });

    if (response.ok) {
      const data = await response.json();
      localStorage.setItem("accessToken", data.token);
      this.reset();
      window.location.href = '/';
    } else {
      console.log('Sign-in failed with status: ' + response.status);
    }
  } catch (err) {
    console.log('Error: ' + err.message);
  }
});
