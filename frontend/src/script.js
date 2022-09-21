function shortenUrl() {
  const input = document.getElementById('url-input')
  const url = input.value
  const data = {url}

  fetch('http://localhost:8080/shorten', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(data),
  })
    .then(response => response.json())
    .then(data => {
      const shortenedUrl = document.getElementById('shortened-url')
      shortenedUrl.innerHTML = `http://${data.short_url}`
      shortenedUrl.setAttribute('href', `http://${data.short_url}`)
    })
    .catch(error => {
      console.error('Error:', error)
    })
}
