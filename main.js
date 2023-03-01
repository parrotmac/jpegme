import React, {useEffect, useState} from 'react';
import ReactDOM from 'react-dom/client';

function App() {
    const [loading, setLoading] = useState(false);
    const [quality, setQuality] = useState(10);
    const [iterations, setIterations] = useState(1);
    const [imageUrl, setImageUrl] = useState("");
    const [b64Image, setB64Image] = useState("");
    const [imgData, setImgData] = useState(new ArrayBuffer(0));

    useEffect(() => {
        fetch('/api/convert', {
            method: "POST",
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify({
                quality: parseInt(quality),
                iterations: parseInt(iterations),
                image: imgData,
            }),
        }).then((body) => {
            body.text().then(data =>
            {
                console.log(data);
                setB64Image(`data:image/jpeg;base64,${data}`)
            })
        })
    }, [imgData, quality])

    return (
        <div className="row justify-content-center">
            <div className="col justify-content-center">
                <div className="card">
                    <h5 className="card-header">JPEG ME</h5>
                    <div className="card-body d-flex flex-column justify-content-center">
                        <div className="input-group mb-3">
                            <span className="input-group-text" id="image-url-label">Image URL</span>
                            <input
                                type={"text"}
                                className={"form-control"}
                                placeholder={'https://example.com/giant-spaghetti.jpg'}
                                aria-describedby="image-url-label"
                                value={imageUrl}
                                onChange={(e) => {
                                    setLoading(true);
                                    setB64Image("");
                                    const url = e?.target?.value;
                                    try {
                                        new URL(url);
                                        setImageUrl(url)
                                    } catch (_) {
                                    }
                                }}/>
                        </div>
                        <span>- or -</span>
                        <input type={'file'} accept={'image/*'} onChange={(e) => {
                            if (e.target.files.length < 1) {
                                return;
                            }
                            const fr = new FileReader();
                            fr.onload = function () {
                                setImgData(fr.result);
                            };
                            fr.readAsDataURL(e.target.files[0]);
                        }} />
                        <hr />
                        <div className="mb-3">
                            <span className="input-group" id="image-quality-label">Quality</span>
                            <input type={'range'}
                                   min={'1'}
                                   max={'20'}
                                   value={quality}
                                   aria-describedby="image-quality-label"
                                   onChange={(e) => {
                                       setLoading(true);
                                       setQuality(e?.target?.value || 1)
                                   }}/>
                        </div>
                        <div className="mb-3">
                            <span className="input-group" id="encode-iterations-label">Iterations</span>
                            <input type={'range'}
                                   min={'1'}
                                   max={'20'}
                                   value={iterations}
                                   aria-describedby="encode-iterations-label"
                                   onChange={(e) => {
                                       setLoading(true);
                                       setIterations(e?.target?.value || 1)
                                   }}/>
                        </div>
                        <p
                            style={{
                                display: loading ? "block" : "none",
                            }}
                        >Loading...</p>
                        <img
                            style={{
                                display: loading ? "none" : "block",
                            }}
                            alt={'Some Distorted Image ¯\\_(ツ)_/¯'}
                            src={
                                (b64Image && b64Image.length > 0) ? b64Image :
                            `/api/convert?quality=${quality}&iterations=${iterations}&image_url=${encodeURIComponent(imageUrl || '')}`
                        }
                            onLoad={() => setLoading(false)}
                        />
                    </div>
                </div>
            </div>
        </div>
    )
}

const container = document.getElementById('root');

const root = ReactDOM.createRoot(container);

root.render(<App/>);