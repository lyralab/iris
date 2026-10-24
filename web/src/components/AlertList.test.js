import React, { act } from 'react';
import { createRoot } from 'react-dom/client';
import AlertList from './AlertList';

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
