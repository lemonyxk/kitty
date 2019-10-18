window.onload = () => {

    setInterval(() => {
        document.querySelector("#timer").innerHTML = Math.round((Date.now() / 100)).toString()
    },100)

}