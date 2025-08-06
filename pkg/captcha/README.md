# Captcha Service Documentation

## Overview

This is a simple captcha service that generates basic math problems (addition and subtraction) for user verification. It's designed to be lightweight and easy to integrate.

## Features

- **Simple Math Captchas**: Generates addition and subtraction problems
- **Secure Random Generation**: Uses cryptographically secure random number generation
- **Automatic Cleanup**: Expired captchas are automatically removed
- **One-time Use**: Each captcha can only be verified once
- **Configurable Expiry**: Default 5-minute expiry (configurable)

## API Endpoints

### Generate Captcha
```
GET /v0/captcha/generate
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "id": "A1B2C3D4E5F6G7H8",
    "question": "15 + 7 = ?"
  }
}
```

### Verify Captcha
```
POST /v0/captcha/verify
Content-Type: application/json
```

**Request Body:**
```json
{
  "id": "A1B2C3D4E5F6G7H8",
  "answer": 22
}
```

**Response (Valid):**
```json
{
  "status": "success",
  "valid": true
}
```

**Response (Invalid):**
```json
{
  "status": "success",
  "valid": false
}
```

## Usage Example

### Frontend Integration (JavaScript)

```javascript
// Generate captcha
async function generateCaptcha() {
  const response = await fetch('/v0/captcha/generate');
  const data = await response.json();
  
  if (data.status === 'success') {
    document.getElementById('captcha-question').textContent = data.data.question;
    document.getElementById('captcha-id').value = data.data.id;
  }
}

// Verify captcha
async function verifyCaptcha(captchaId, answer) {
  const response = await fetch('/v0/captcha/verify', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      id: captchaId,
      answer: parseInt(answer)
    })
  });
  
  const data = await response.json();
  return data.valid;
}
```

### HTML Form Example

```html
<form id="signup-form">
  <!-- Other form fields -->
  
  <div class="captcha-section">
    <label for="captcha-answer">
      Please solve: <span id="captcha-question">Loading...</span>
    </label>
    <input type="hidden" id="captcha-id" name="captcha-id" />
    <input type="number" id="captcha-answer" name="captcha-answer" required />
    <button type="button" onclick="generateCaptcha()">Refresh Captcha</button>
  </div>
  
  <button type="submit">Sign Up</button>
</form>

<script>
// Generate initial captcha when page loads
window.onload = function() {
  generateCaptcha();
};

// Handle form submission
document.getElementById('signup-form').onsubmit = async function(e) {
  e.preventDefault();
  
  const captchaId = document.getElementById('captcha-id').value;
  const captchaAnswer = document.getElementById('captcha-answer').value;
  
  const isValidCaptcha = await verifyCaptcha(captchaId, captchaAnswer);
  
  if (!isValidCaptcha) {
    alert('Invalid captcha answer. Please try again.');
    generateCaptcha(); // Generate new captcha
    return;
  }
  
  // Proceed with form submission
  // ... rest of your form handling code
};
</script>
```

## Security Features

1. **Cryptographically Secure Random Numbers**: Uses `crypto/rand` for generating random numbers
2. **One-time Use**: Each captcha is deleted after verification
3. **Automatic Expiry**: Captchas expire after 5 minutes
4. **Thread-safe**: Uses mutexes for concurrent access
5. **Memory Efficient**: Automatic cleanup prevents memory leaks

## Configuration

The captcha service can be customized:

```go
// Create service with custom expiry
service := captcha.NewCaptchaService()
service.SetExpiry(10 * time.Minute) // 10 minutes instead of default 5
```

## Testing

Run the tests with:
```bash
go test ./pkg/captcha/...
```

## Integration with User Registration

You can integrate this with your user registration by adding captcha verification before creating users:

```go
// In your user registration handler
func RegisterUser(c *gin.Context) {
    // ... get user data and captcha verification ...
    
    // Verify captcha first
    isValidCaptcha := captchaService.VerifyCaptcha(&captcha.CaptchaVerification{
        ID:     captchaData.ID,
        Answer: captchaData.Answer,
    })
    
    if !isValidCaptcha {
        c.JSON(400, gin.H{"error": "Invalid captcha"})
        return
    }
    
    // Proceed with user creation...
}
```
