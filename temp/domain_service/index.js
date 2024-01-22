require('dotenv').config()
const express = require("express");

const app = express()

app.get("/", (req, res) => {
    console.log("Requester IP: ", req.ip);
    return res.status(200).send("<h1>Hello World</h1>")
})

app.listen(process.env.SERVICE_PORT, () => {
    console.log(`Service is running on: ${process.env.SERVICE_PORT}`);
})

// 4949
// 9494
// // 9994
// 9944