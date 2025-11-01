// const base_url= 'http://127.0.0.1:9090' ;
//
// const config = {
//     api: {
//         alertSummary: base_url + '/v0/alerts/firingCount',
//         firingIssues: base_url + '/v0/alerts/?page=1&pagination=10&status=firing',
//         resolvedIssues: base_url + '/v0/alerts/?page=1&pagination=10&status=resolved',
//     },
//     auth: {
//         username: 'admin', // Replace with your actual username
//         password: '123',  // Replace with your actual password
//     },
// };
//
// export default config;
const base_url = 'http://127.0.0.1:9090';

const config = {
    api: {
        // Health endpoints
        health: base_url + '/healthy',
        ready: base_url + '/ready',

        // Captcha endpoints
        captchaGenerate: base_url + '/v0/captcha/generate',

        // Auth endpoints
        signin: base_url + '/v0/users/signin',

        // Alert endpoints
        alerts: base_url + '/v0/alerts',
        alertSummary: base_url + '/v0/alerts/firingCount',
        firingIssues: base_url + '/v0/alerts?status=firing',
        resolvedIssues: base_url + '/v0/alerts?status=resolved',

        // User endpoints
        users: base_url + '/v0/users',
        userMe: base_url + '/v0/users/me',
        userVerify: base_url + '/v0/users/verify',
        userGroups: (userId) => base_url + `/v0/users/${userId}/groups`,

        // Group endpoints
        groups: base_url + '/v0/groups',
        group: (groupId) => base_url + `/v0/groups/${groupId}`,
        groupUsers: (groupId) => base_url + `/v0/groups/${groupId}/users`,
        addUserToGroup: (groupId) => base_url + `/v0/groups/${groupId}/users`,

        // Provider endpoints
        providers: base_url + '/v0/providers',

        // Message endpoints
        alertManagerMessage: base_url + '/v1/messages/alertmanager',
    },
};

export default config;