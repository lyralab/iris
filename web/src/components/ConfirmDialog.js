import Icon from './Icon';
import './ConfirmDialog.css';

const ConfirmDialog = ({ title, message, confirmLabel, danger = true, onConfirm, onCancel }) => (
    <div className="confirm-overlay" onClick={onCancel}>
        <div
            className="modal confirm-dialog"
            role="dialog"
            aria-modal="true"
            aria-labelledby="confirm-dialog-title"
            onClick={(e) => e.stopPropagation()}
        >
            <div className="confirm-body">
                <span className={`confirm-icon ${danger ? 'danger' : ''}`}>
                    <Icon name={danger ? 'alert' : 'checkCircle'} size={20} />
                </span>
                <div>
                    <h2 id="confirm-dialog-title">{title}</h2>
                    <p>{message}</p>
                </div>
            </div>
            <div className="modal-footer">
                <button type="button" className="btn-secondary" onClick={onCancel}>Cancel</button>
                <button type="button" className={danger ? 'btn-danger' : 'btn-primary'} onClick={onConfirm} autoFocus>{confirmLabel}</button>
            </div>
        </div>
    </div>
);

export default ConfirmDialog;
