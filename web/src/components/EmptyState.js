import React from 'react';
import Icon from './Icon';

const EmptyState = ({ icon = 'inbox', title, children, tone = '' }) => (
    <div className={`empty-state ${tone}`.trim()}>
        <span className="empty-state-icon"><Icon name={icon} size={20} /></span>
        {title && <strong>{title}</strong>}
        {children && <p>{children}</p>}
    </div>
);

export default EmptyState;
