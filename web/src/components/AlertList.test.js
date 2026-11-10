import React, { act } from 'react';
import { createRoot } from 'react-dom/client';
import AlertList, { isSilenced } from './AlertList';
import { isAdmin } from '../utils/auth';
import apiService from '../utils/apiService';

jest.mock('../utils/auth', () => ({ isAdmin: jest.fn(() => false) }));
jest.mock('../utils/apiService', () => ({ silenceAlert: jest.fn(), unsilenceAlert: jest.fn() }));

window.IS_REACT_ACT_ENVIRONMENT = true;

let container;
let root;

beforeEach(() => {
    container = document.createElement('div');
    document.body.appendChild(container);
    root = createRoot(container);
});

afterEach(() => {
    act(() => root.unmount());
    container.remove();
});

const kubeAlert = {
    id: 'a1',
    name: 'KubeJobCompletion',
    severity: 'warning',
    status: 'firing',
    description: 'Job did not complete in time',
    starts_at: '2026-10-24T03:30:24Z',
    labels: {
        alertname: 'KubeJobCompletion',
        severity: 'warning',
        receptor: 'ops',
        namespace: 'batch',
        job_name: 'nightly-report',
    },
    annotations: {
        summary: 'Job did not complete in time',
        description: 'Job batch/nightly-report is taking more than 12 hours to complete.',
        runbook_url: 'https://runbooks.example/kubejobcompletion',
        dashboard: 'jobs',
        grafana_url: 'https://grafana.example/d/jobs',
    },
    // eslint-disable-next-line no-script-url -- checking that this URL is not linked
    generator_url: 'javascript:alert(1)',
};

const renderExpanded = (alert) => {
    // eslint-disable-next-line testing-library/no-unnecessary-act -- raw react-dom, no RTL installed
    act(() => root.render(<AlertList alerts={[alert]} />));
    act(() => container.querySelector('.alert-row-main').click());
};

test('expanded row shows labels, annotations and the runbook link', () => {
    renderExpanded(kubeAlert);

    expect(container.querySelector('.alert-row-fulldesc').textContent)
        .toBe('Job batch/nightly-report is taking more than 12 hours to complete.');
    const chips = [...container.querySelectorAll('.label-chip')].map((li) => li.textContent);
    expect(chips).toEqual(['job_name=nightly-report', 'namespace=batch']);
    expect(container.textContent).toContain('dashboard');
    const links = [...container.querySelectorAll('.alert-row-links a')].map((a) => [a.textContent, a.getAttribute('href')]);
    expect(links).toEqual([
        ['Runbook', 'https://runbooks.example/kubejobcompletion'],
        ['Grafana', 'https://grafana.example/d/jobs'],
    ]);
    // URL annotations are links, not plain fields.
    expect([...container.querySelectorAll('dt')].map((dt) => dt.textContent)).not.toContain('grafana url');
});

test('non-http generator URLs are not linked', () => {
    renderExpanded(kubeAlert);

    expect(container.textContent).not.toContain('View in Prometheus');
});

test('alerts without labels render without a labels section', () => {
    renderExpanded({ ...kubeAlert, labels: undefined, annotations: undefined, generator_url: undefined });

    expect(container.querySelector('.alert-row-labels')).toBeNull();
    expect(container.querySelector('.alert-row-fulldesc').textContent).toBe('Job did not complete in time');
});

const silenceButton = () => [...container.querySelectorAll('.alert-row-silence button')][0];

test('viewers see the silenced badge but no silence controls', () => {
    isAdmin.mockReturnValue(false);
    renderExpanded({ ...kubeAlert, silenced: true, silenced_by: 'root' });

    expect(container.querySelector('.silenced-badge')).not.toBeNull();
    expect(container.textContent).toContain('Notifications off for everyone by root, until unsilenced');
    expect(container.querySelector('.alert-row-silence')).toBeNull();
});

test('admin silences a firing alert with the chosen duration', async () => {
    isAdmin.mockReturnValue(true);
    apiService.silenceAlert.mockResolvedValue({ alert: { silenced: true, silenced_by: 'root' } });
    renderExpanded(kubeAlert);
    expect(container.querySelector('.silenced-badge')).toBeNull();

    const select = container.querySelector('.alert-row-silence select');
    act(() => {
        select.value = '24h';
        select.dispatchEvent(new Event('change', { bubbles: true }));
    });
    await act(async () => silenceButton().click());

    expect(apiService.silenceAlert).toHaveBeenCalledWith('a1', '24h');
    expect(container.querySelector('.silenced-badge')).not.toBeNull();
    expect(silenceButton().textContent).toContain('Unsilence');
});

test('admin unsilences a silenced alert', async () => {
    isAdmin.mockReturnValue(true);
    apiService.unsilenceAlert.mockResolvedValue({ alert: { silenced: false } });
    renderExpanded({ ...kubeAlert, silenced: true });

    await act(async () => silenceButton().click());

    expect(apiService.unsilenceAlert).toHaveBeenCalledWith('a1');
    expect(container.querySelector('.silenced-badge')).toBeNull();
});

test('admin sees no silence control on a resolved, unsilenced alert', () => {
    isAdmin.mockReturnValue(true);
    renderExpanded({ ...kubeAlert, status: 'resolved' });

    expect(container.querySelector('.alert-row-silence')).toBeNull();
});

test('isSilenced drops an expired silence', () => {
    const now = Date.parse('2026-11-10T12:00:00Z');
    expect(isSilenced({ silenced: true }, now)).toBe(true);
    expect(isSilenced({ silenced: true, silenced_until: '2026-11-10T13:00:00Z' }, now)).toBe(true);
    expect(isSilenced({ silenced: true, silenced_until: '2026-11-10T11:00:00Z' }, now)).toBe(false);
    expect(isSilenced({ silenced: false }, now)).toBe(false);
});
