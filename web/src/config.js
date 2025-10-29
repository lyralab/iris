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
        // Alert endpoints
        alertSummary: base_url + '/v0/alerts/firingCount',
        firingIssues: base_url + '/v0/alerts/?page=1&pagination=10&status=firing',
        resolvedIssues: base_url + '/v0/alerts/?page=1&pagination=10&status=resolved',
        
        // User endpoints
        signin: base_url + '/v0/users/signin',
        users: base_url + '/v0/users',
        userInfo: base_url + '/v0/users/me',
        verifyUser: base_url + '/v0/users/verify',
        
        // Group endpoints
        groups: base_url + '/v0/groups',
        userGroups: (userId) => base_url + `/v0/users/${userId}/groups`,
        groupUsers: (groupId) => base_url + `/v0/groups/${groupId}/users`,
        addUserToGroup: (groupId) => base_url + `/v0/groups/${groupId}/users`,
        
        // Captcha endpoint
        captcha: base_url + '/v0/captcha',
    },
};

export default config;