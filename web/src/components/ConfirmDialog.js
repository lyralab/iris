import './ConfirmDialog.css';

const ConfirmDialog = ({ title, message, confirmLabel, danger = true, onConfirm, onCancel }) => (
    <div className="confirm-overlay" onClick={onCancel}>
        <div
            className="confirm-dialog"
            role="dialog"
            aria-modal="true"
            aria-labelledby="confirm-dialog-title"
            onClick={(e) => e.stopPropagation()}
        >
            <h2 id="confirm-dialog-title">{title}</h2>
            <p>{message}</p>
            <div className="confirm-actions">
                <button type="button" className="confirm-cancel" onClick={onCancel}>Cancel</button>
                <button type="button" className={danger ? 'confirm-danger' : 'confirm-ok'} onClick={onConfirm}>{confirmLabel}</button>
            </div>
        </div>
    </div>
);

export default ConfirmDialog;
