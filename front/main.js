window.addEventListener('DOMContentLoaded', (event) => {
  var latestTimer
  document.getElementById('input-box').addEventListener('input', (event) => {
    let currentTimer = setTimeout(() => {
      const targetValue = event.target.value
      document.getElementById('input-value').textContent = targetValue
    }, 5000)
    if (latestTimer) {
      clearTimeout(latestTimer)
    }
    latestTimer = currentTimer
  })

  document.getElementById('fetch-from-server').addEventListener('click', async (event) => {
    console.log('fetching from server...')
    const response = await fetch('http://localhost:8888').then(res => {console.log(res);return res})
    // const response = await fetch('http://localhost:8888').then(res => {console.log(res);return res.json()}) → json()メソッドを実行した結果としてPromiseが返却されるが、awaitされていることで中身を取り出して変数に格納している
    .catch(err => console.log(err))
    console.log(await response.json()) // json()メソッドはPromiseを返す == awaitすることで値を取り出せる
  })
})
