import './styles/index.css';

import * as React from 'react';
import ReactDOM from 'react-dom/client';

import App from '@/app';

ReactDOM.createRoot(document.querySelector('#root')!).render(
    <React.StrictMode>
        <App />
    </React.StrictMode>,
);
