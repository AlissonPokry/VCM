const errorCodes = {
  400: 'BAD_REQUEST',
  401: 'UNAUTHORIZED',
  404: 'NOT_FOUND',
  409: 'CONFLICT',
  422: 'VALIDATION_ERROR',
  500: 'INTERNAL_SERVER_ERROR'
};

class AppError extends Error {
  constructor(message, status = 500, code = null) {
    super(message);
    this.status = status;
    this.code = code || errorCodes[status] || 'INTERNAL_SERVER_ERROR';
  }
}

function errorHandler(err, req, res, next) {
  if (res.headersSent) {
    return next(err);
  }

  const status = err.status || 500;
  const message = status === 500 && process.env.NODE_ENV === 'production'
    ? 'Internal server error'
    : err.message || 'Internal server error';

  if (status >= 500) {
    console.error(err);
  }

  return res.status(status).json({
    error: true,
    message,
    code: err.code || errorCodes[status] || 'INTERNAL_SERVER_ERROR'
  });
}

module.exports = { AppError, errorHandler };
