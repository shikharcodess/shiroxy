require('dotenv').config()
const express = require("express");

const app = express()

app.get("/", (req, res) => {
    console.log("Requester IP: ", req.ip);

    return res.status(200).send(`<h1>Hello World</h1><h1>Running on port ${process.argv[2]}</h1>`)
})

app.listen(process.argv[2], () => {
    console.log(`Service is running on: ${process.argv[2]}`);
})

// 4949
// 9494
// // 9994
// 9944