import React, {useState} from 'react';
import ReactDOM from 'react-dom/client';

function App() {
    const [loading, setLoading] = useState(false);
    const [quality, setQuality] = useState(10);
    const [imageUrl, setImageUrl] = useState("");

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
                                    const url = e?.target?.value;
                                    try {
                                        new URL(url);
                                        setImageUrl(url)
                                    } catch (_) {
                                    }
                                }}/>
                        </div>
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
                            src={`/api/convert?quality=${quality}&image_url=${encodeURIComponent(imageUrl || '')}`}
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