package public

var DOMAIN_NOT_FOUND_ERROR string = `<!DOCTYPE html>"
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>404 Not Found</title>
    <style>
      * {
        box-sizing: border-box;
        margin: 0;
        padding: 0;
      }

      body,
      html {
        height: 100%;
        font-family: "Arial", sans-serif;
        background: #080e1f; /* Light grey background */
        display: flex;
        justify-content: center;
        align-items: center;
        text-align: center;
      }

      .container {
        padding: 20px;
      }

      .error-image-div {
        display: flex;
        justify-content: center;
        height: 50vh;
      }

      .shiroxy-logo-div {
        display: flex;
        justify-content: center;
        height: 40%;
      }

      .shiroxy-logo {
        width: 40%;
        height: 40%;
        /* height: auto; */
      }

      .error-image {
        width: 80%;
        height: 50%;
        height: auto;
      }

      .main-heading {
        font-size: 2em;
        color: #ffffff; /* Dark grey color */
        margin-bottom: 10px;
      }

      .sub-heading {
        font-size: 1.5em;
        color: #666; /* Medium grey color */
        margin-bottom: 20px;
      }

      .info-text {
        color: #888; /* Light grey color */
        margin-bottom: 30px;
      }

      .button {
        display: inline-block;
        padding: 10px 20px;
        font-size: 1em;
        border: none;
        border-radius: 5px;
        background: #1D56C4; /* Blue background */
        color: white;
        text-decoration: none;
        transition: background 0.3s;
      }

      .button:hover {
        background: #4b6cb7; /* Darker blue on hover */
      }

      @media screen and (max-width: 600px) {
        .error-image-div {
          height: 300px; /* Adjust the width as needed */
          margin: 0 auto; /* Center the element horizontally */
          overflow: hidden;
        }

        .error-image {
          height: 300px;
        }
      }
    </style>
  </head>
  <body>
    <div class="container">

        <div class="shiroxy-logo-div">
            <img
              class="shiroxy-logo"
              src="/home/shikharcode/Main/opensource/shiroxy/media/shiroxy_logo.png"
              alt="Shiroxy Logo"
            />
        </div>

      <!-- Replace with your own SVG code or use an img tag -->
      <div class="error-image-div">
        <img
          class="error-image"
          src="/home/shikharcode/Main/opensource/shiroxy/public/404_new.svg"
          alt="Centered SVG"
        />
      </div>

      <h1 class="main-heading">Oops! Page not found.</h1>
      <h2 class="sub-heading">
        We can't seem to find the page you're looking for.
      </h2>
      <p class="info-text">
        The page you are looking for might have been removed, had its name
        changed, or is temporarily unavailable.
      </p>
      <a href="{{button_url}}" target="_blank" class="button">{{button_name}}</a>
    </div>
  </body>
</html>`
