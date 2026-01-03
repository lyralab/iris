import config from '../config';
import { getAuthToken } from './auth';

/**
 * API Service - Centralized API call handler
 * Handles all HTTP requests with authentication
 */

class APIService {
    /**
     * Generic fetch wrapper with auth handling
     */
    async fetch(url, options = {}) {
        const token = getAuthToken();
        const headers = {
            'Content-Type': 'application/json',
            ...options.headers,
        };

        if (token) {
            headers['Authorization'] = `Bearer ${token}`;
        }

        const response = await fetch(url, {
            ...options,
            headers,
        });

        if (!response.ok) {
            const error = await response.json().catch(() => ({ message: 'Request failed' }));
            const errorMessage = error.message || error.error || `HTTP ${response.status}`;
            throw new Error(errorMessage);
        }

        return response.json();
    }

    // ============ Health Endpoints ============
    async checkHealth() {
        return this.fetch(config.api.health);
    }

    async checkReady() {
        return this.fetch(config.api.ready);
    }

    // ============ Captcha Endpoints ============
    async generateCaptcha() {
        return this.fetch(config.api.captchaGenerate);
    }

    // ============ Auth Endpoints ============
    async signin(username, password, captchaId, captchaAnswer) {
        return this.fetch(`${config.api.signin}?captcha_id=${captchaId}`, {
            method: 'POST',
            body: JSON.stringify({ username, password, captcha_answer: captchaAnswer }),
        });
    }

    // ============ User Endpoints ============
    async getUsers() {
        return this.fetch(config.api.users);
    }

    async addUser(userData) {
        return this.fetch(config.api.users, {
            method: 'POST',
            body: JSON.stringify(userData),
        });
    }

    async updateUser(userData) {
        return this.fetch(config.api.users, {
            method: 'PUT',
            body: JSON.stringify(userData),
        });
    }

    async verifyUser(username) {
        return this.fetch(config.api.userVerify, {
            method: 'PUT',
            body: JSON.stringify({ username }),
        });
    }

    async getCurrentUser() {
        return this.fetch(config.api.userMe);
    }

    async getUserMe() {
        return this.getCurrentUser();
    }

    async getUserGroups(userId) {
        return this.fetch(config.api.userGroups(userId));
    }

    // ============ Group Endpoints ============
    async getGroups() {
        return this.fetch(config.api.groups);
    }

    async getGroup(groupId) {
        return this.fetch(`${config.api.groups}?id=${groupId}`);
    }

    async createGroup(groupData) {
        return this.fetch(config.api.groups, {
            method: 'POST',
            body: JSON.stringify(groupData),
        });
    }

    async deleteGroup(groupData) {
        return this.fetch(config.api.groups, {
            method: 'DELETE',
            body: JSON.stringify(groupData),
        });
    }

    async getGroupUsers(groupId) {
        return this.fetch(config.api.groupUsers(groupId), {
            headers: {
                'Content-Type': 'application/json',
            },
        });
    }

    async addUserToGroup(groupId, userId) {
        return this.fetch(config.api.addUserToGroup(groupId), {
            method: 'POST',
            body: JSON.stringify({ user_id: userId }),
        });
    }

    // ============ Alert Endpoints ============
    async getAlerts(params = {}) {
        const queryParams = new URLSearchParams();
        if (params.status) queryParams.append('status', params.status);
        if (params.severity) queryParams.append('severity', params.severity);
        if (params.limit) queryParams.append('limit', params.limit);
        if (params.page) queryParams.append('page', params.page);
        
        const url = `${config.api.alerts}?${queryParams.toString()}`;
        return this.fetch(url);
    }

    async getAlertSummary() {
        return this.fetch(config.api.alertSummary);
    }

    async getFiringAlerts(limit = 10, page = 1) {
        return this.fetch(`${config.api.firingIssues}&limit=${limit}&page=${page}`);
    }

    async getResolvedAlerts(limit = 10, page = 1) {
        return this.fetch(`${config.api.resolvedIssues}&limit=${limit}&page=${page}`);
    }

    // ============ Provider Endpoints ============
    async getProviders() {
        return this.fetch(config.api.providers);
    }

    async getProvider(identifier, type = 'name') {
        const param = type === 'id' ? 'id' : 'name';
        return this.fetch(`${config.api.providers}?${param}=${identifier}`);
    }

    async modifyProvider(providerData) {
        return this.fetch(config.api.providers, {
            method: 'PUT',
            body: JSON.stringify(providerData),
        });
    }

    // ============ Message Endpoints ============
    async sendAlertManagerMessage(messageData) {
        return this.fetch(config.api.alertManagerMessage, {
            method: 'POST',
            body: JSON.stringify(messageData),
            headers: {
                'Authorization': 'Basic ' + btoa('admin:admin'), // Basic auth for messages
            },
        });
    }
}

// Export singleton instance
const apiService = new APIService();
export default apiService;

