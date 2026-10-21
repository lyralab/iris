// jsdom in CRA's Jest lacks TextEncoder, which react-router 7 needs at import time.
import { TextEncoder, TextDecoder } from 'util';

if (typeof global.TextEncoder === 'undefined') {
    global.TextEncoder = TextEncoder;
    global.TextDecoder = TextDecoder;
}
