import React, { useEffect, useState } from "react";
import ReactDOM from "react-dom/client";

function App() {
  const [loading, setLoading] = useState(false);
  const [quality, setQuality] = useState(10);
  const [interleaveGif, setInterleaveGif] = useState(false);
  const [iterations, setIterations] = useState(1);

  const [errorMsg, setErrorMessage] = useState("");
  const [imageData, setImageData] = useState(
    "https://isaacparker.org/SunChips.jpg"
  );

  // Either a fully-qualified URL or a data URL
  const [displayImage, setDisplayImage] = useState("");

  useEffect(() => {
    const reqOptions = {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        params: {
          quality: quality,
          iterations: iterations,
          interleave_gif: interleaveGif,
        },
        image: imageData, // Either url raw image
      }),
    };

    fetch("/api/convert", reqOptions).then((resp) => {
      resp.text().then((body) => {
        if (resp.status === 200) {
          setDisplayImage(`data:image/jpeg;base64,${body}`);
          setErrorMessage("");
        } else {
          setDisplayImage("");
          setErrorMessage(body);
        }
        setLoading(false);
      });
    });
  }, [quality, iterations, imageData]);

  return (
    <div className="row justify-content-center">
      <div className="col justify-content-center">
        <div className="card">
          <h5 className="card-header">JPEG ME</h5>
          <div className="card-body d-flex flex-column justify-content-center">
            <div className="input-group mb-3">
              <span className="input-group-text" id="image-url-label">
                Image URL
              </span>
              <input
                type={"text"}
                className={"form-control"}
                placeholder={"https://example.com/giant-spaghetti.jpg"}
                aria-describedby="image-url-label"
                value={imageData}
                onChange={(e) => {
                  setLoading(true);
                  const url = e?.target?.value;
                  try {
                    new URL(url);
                    setImageData(url);
                  } catch (_) {}
                }}
              />
            </div>
            <span>- or -</span>
            <input
              type={"file"}
              accept={"image/*"}
              onChange={(e) => {
                if (e.target.files.length < 1) {
                  return;
                }
                const fr = new FileReader();
                fr.onload = function () {
                  setImageData(fr.result.toString());
                };
                fr.readAsDataURL(e.target.files[0]);
              }}
            />
            <hr />
            <div className="mb-3">
              <span className="input-group" id="image-quality-label">
                Quality
              </span>
              <input
                type={"range"}
                min={"1"}
                max={"20"}
                value={quality}
                aria-describedby="image-quality-label"
                onChange={(e) => {
                  setLoading(true);
                  setQuality(parseInt(e?.target?.value || 1));
                }}
              />
            </div>
            <div className="mb-3">
              <span className="input-group" id="encode-iterations-label">
                Iterations
              </span>
              <input
                type={"range"}
                min={"1"}
                max={"20"}
                value={iterations}
                aria-describedby="encode-iterations-label"
                onChange={(e) => {
                  setLoading(true);
                  setIterations(parseInt(e?.target?.value || 1));
                }}
              />
            </div>
            <div className="mb-3">
              <span className="input-group" id="encode-gif-label">
                Interleave GIF Encoding
              </span>
              <input
                type={"checkbox"}
                checked={interleaveGif}
                aria-describedby="encode-gif-label"
                onChange={(e) => {
                  setLoading(true);
                  setInterleaveGif(!interleaveGif);
                }}
              />
            </div>
            <p
              style={{
                display: loading ? "block" : "none",
              }}
            >
              Loading...
            </p>
            <p>{errorMsg}</p>
            <img
              style={{
                display:
                  loading || !displayImage || displayImage.length === 0
                    ? "none"
                    : "block",
              }}
              alt={"Some Distorted Image ¯\\_(ツ)_/¯"}
              src={displayImage}
            />
          </div>
        </div>
      </div>
    </div>
  );
}

const container = document.getElementById("root");

const root = ReactDOM.createRoot(container);

root.render(<App />);
