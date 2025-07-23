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
      console.log('Registered successfully! Signing in...');
      const signInResponse = await fetch('/api/auth/sign-in', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password })
      });

      if (signInResponse.ok) {
        const data = await signInResponse.json();
        localStorage.setItem("accessToken", data.token);
        this.reset();
        window.location.href = '/';
      } else {
        console.log('Sign-in after registration failed with status: ' + signInResponse.status);
      }
    } else {
      console.log('Registration failed with status: ' + response.status);
    }
  } catch (err) {
    console.log('Error: ' + err.message);
  }
});