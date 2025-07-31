import React, { useEffect, useState } from 'react';

interface RateLimitEvent extends CustomEvent {
  detail: {
    message: string;
    retryAfter: number;
  };
}

export default function RateLimitNotification() {
  const [show, setShow] = useState(false);
  const [message, setMessage] = useState('');
  const [countdown, setCountdown] = useState(0);

  useEffect(() => {
    const handleRateLimit = (event: Event) => {
      const rateLimitEvent = event as RateLimitEvent;
      setMessage(rateLimitEvent.detail.message);
      setCountdown(rateLimitEvent.detail.retryAfter);
      setShow(true);
    };

    window.addEventListener('rate-limit-error', handleRateLimit);
    return () => window.removeEventListener('rate-limit-error', handleRateLimit);
  }, []);

  useEffect(() => {
    if (countdown > 0) {
      const timer = setTimeout(() => {
        setCountdown(countdown - 1);
      }, 1000);
      return () => clearTimeout(timer);
    } else if (countdown === 0 && show) {
      // Hide notification when countdown reaches 0
      setTimeout(() => setShow(false), 2000);
    }
  }, [countdown, show]);

  if (!show) return null;

  return (
    <div className="fixed top-4 right-4 max-w-md bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded shadow-lg z-50">
      <div className="flex items-center">
        <svg className="w-6 h-6 mr-2" fill="currentColor" viewBox="0 0 20 20">
          <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
        </svg>
        <div>
          <p className="font-bold">Rate Limit Exceeded</p>
          <p className="text-sm">{message}</p>
          {countdown > 0 && (
            <p className="text-sm mt-1">Retry in: {countdown} seconds</p>
          )}
        </div>
      </div>
    </div>
  );
}