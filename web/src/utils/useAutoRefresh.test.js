import { skipWhileRunning, latestOnly } from './useAutoRefresh';

test('skipWhileRunning drops calls while the previous one is pending', async () => {
    let release;
    const fn = jest.fn(() => new Promise((resolve) => { release = resolve; }));
    const tick = skipWhileRunning(fn);

    const first = tick();
    tick();
    tick();
    expect(fn).toHaveBeenCalledTimes(1);

    release();
    await first;
    tick();
    expect(fn).toHaveBeenCalledTimes(2);
});

test('skipWhileRunning frees the slot after a failure', async () => {
    const fn = jest.fn().mockRejectedValueOnce(new Error('boom')).mockResolvedValue();
    const tick = skipWhileRunning(fn);

    await expect(tick()).rejects.toThrow('boom');
    await tick();
    expect(fn).toHaveBeenCalledTimes(2);
});

test('latestOnly marks earlier requests stale once a newer one starts', () => {
    const start = latestOnly();
    const a = start();
    expect(a()).toBe(true);
    const b = start();
    expect(a()).toBe(false);
    expect(b()).toBe(true);
});
