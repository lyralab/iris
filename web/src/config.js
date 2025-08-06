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
const base_url= 'http://127.0.0.1:9090' ;

const config = {
    api: {
        alertSummary: base_url + '/v0/alerts/firingCount',
        firingIssues: base_url + '/v0/alerts/?page=1&pagination=10&status=firing',
        resolvedIssues: base_url + '/v0/alerts/?page=1&pagination=10&status=resolved',
        signin: base_url + '/v1/users/signin',
    },
};

export default config;