import React, { useEffect } from 'react';
import Icon from './Icon';

// Dialog shell: overlay click and Escape close it. Put content in <div className="modal-body">
// and buttons in <div className="modal-footer">, or wrap both in a <form>.
const Modal = ({ title, subtitle, onClose, size = '', children }) => {
    useEffect(() => {
        const onKey = (e) => {
            if (e.key === 'Escape') onClose();
        };
        document.addEventListener('keydown', onKey);
        return () => document.removeEventListener('keydown', onKey);
    }, [onClose]);

    return (
        <div className="modal-overlay" onClick={onClose}>
            <div
                className={`modal ${size}`.trim()}
                role="dialog"
                aria-modal="true"
                aria-label={typeof title === 'string' ? title : undefined}
                onClick={(e) => e.stopPropagation()}
            >
                <div className="modal-header">
                    <div>
                        <h2>{title}</h2>
                        {subtitle && <p>{subtitle}</p>}
                    </div>
                    <button type="button" className="btn-ghost btn-icon btn-sm" onClick={onClose} aria-label="Close">
                        <Icon name="x" />
                    </button>
                </div>
                {children}
            </div>
        </div>
    );
};

export default Modal;
