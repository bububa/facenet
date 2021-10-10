<html>
<head>
  <meta charset="utf-8"/>
  <title>facenet</title>
  <style type="text/css">
    html, body {
      margin:0;padding:0;
      background-color: #000;
    } 
    .video-wrapper {
      display:flex;
      justify-content: center;
      align-items: center;
      width: 100vw;
      height: 100vh;
    }
  </style>
</head>
<body>
  <div class="video-wrapper">
    <canvas id="canvas" width="{{ .Width }}" height="{{ .Height }}"></canvas>
  </div>
  <script>
    var canvasElm = document.getElementById("canvas")
    var image = new Image();
    {{ if .UseWebGL }}
	var texture, vloc, tloc, vertexBuff, textureBuff;
    {{ end }}
    connect();
    function connect() {
        var ws = new WebSocket("ws://" + window.location.host + "/socket");
        ws.onopen = function() {
          draw();
        }
        ws.onmessage = function(e) {
          image.setAttribute("src", "data:image/jpeg;base64," + e.data);
        }
        ws.onclose = function(e) {
          console.log('Socket is closed. Reconnect will be attempted in 1 second.', e.reason);
          setTimeout(function() {
            connect();
          }, 1000);
        }
        ws.onerror = function(err) {
          console.error('Socket encountered error: ', err.message, 'Closing socket');
          ws.close();
        }
    }
    {{ if .UseWebGL }}
    function draw() {
      var gl = canvasElm.getContext('webgl',{antialias:false}) || canvas.getContext('experimental-webgl');
      var vertexShaderSrc =
        "attribute vec2 aVertex;" +
        "attribute vec2 aUV;" +
        "varying vec2 vTex;" +
        "void main(void) {" +
        "  gl_Position = vec4(aVertex, 0.0, 1.0);" +
        "  vTex = aUV;" +
        "}";
      var fragmentShaderSrc =
        "precision mediump float;" +
        "varying vec2 vTex;" +
        "uniform sampler2D sampler0;" +
        "void main(void){" +
        "  gl_FragColor = texture2D(sampler0, vTex);"+
        "}";
      var vertShaderObj = gl.createShader(gl.VERTEX_SHADER);
      var fragShaderObj = gl.createShader(gl.FRAGMENT_SHADER);
      gl.shaderSource(vertShaderObj, vertexShaderSrc);
      gl.shaderSource(fragShaderObj, fragmentShaderSrc);
      gl.compileShader(vertShaderObj);
      gl.compileShader(fragShaderObj);
      var program = gl.createProgram();
      gl.attachShader(program, vertShaderObj);
      gl.attachShader(program, fragShaderObj);
      gl.linkProgram(program);
      gl.useProgram(program);
      gl.viewport(0, 0, {{.Width}}, {{.Height}});
      vertexBuff = gl.createBuffer();
      gl.bindBuffer(gl.ARRAY_BUFFER, vertexBuff);
      gl.bufferData(gl.ARRAY_BUFFER, new Float32Array([-1, 1, -1, -1, 1, -1, 1, 1]), gl.STATIC_DRAW);
      textureBuff = gl.createBuffer();
      gl.bindBuffer(gl.ARRAY_BUFFER, textureBuff);
      gl.bufferData(gl.ARRAY_BUFFER, new Float32Array([0, 1, 0, 0, 1, 0, 1, 1]), gl.STATIC_DRAW);
      vloc = gl.getAttribLocation(program, "aVertex");
      tloc = gl.getAttribLocation(program, "aUV");
      texture = gl.createTexture();
      image.onload = function() {
        gl.bindTexture(gl.TEXTURE_2D, texture);
        gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE);
        gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE);
        gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST);
        gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST);
        gl.pixelStorei(gl.UNPACK_FLIP_Y_WEBGL, true);
        gl.texImage2D(gl.TEXTURE_2D, 0,  gl.RGBA,  gl.RGBA, gl.UNSIGNED_BYTE, image);
        gl.bindBuffer(gl.ARRAY_BUFFER, vertexBuff);
        gl.enableVertexAttribArray(vloc);
        gl.vertexAttribPointer(vloc, 2, gl.FLOAT, false, 0, 0);
        gl.bindBuffer(gl.ARRAY_BUFFER, textureBuff);
        gl.enableVertexAttribArray(tloc);
        gl.vertexAttribPointer(tloc, 2, gl.FLOAT, false, 0, 0);
        gl.drawArrays(gl.TRIANGLE_FAN, 0, 4);
      }
    }
    {{ else }}
    function draw() {
      var context = canvasElm.getContext("2d", {alpha: false});
      image.onload = function() {
        var w=image.width;
        var h=image.height;
        var sizer=scalePreserveAspectRatio(w, h, canvasElm.width, canvasElm.height);
        var dw = w * sizer;
        var dh = h * sizer;
        var x = (canvasElm.width - dw) / 2;
        var y = (canvasElm.height - dh) / 2;
        context.drawImage(image, 0, 0, w, h, x, y, dw, dh);
      }
    }
    {{ end }}
    function scalePreserveAspectRatio(imgW,imgH,maxW,maxH){
      return (Math.min((maxW/imgW),(maxH/imgH)));
    }
  </script>
</body>
</html>
