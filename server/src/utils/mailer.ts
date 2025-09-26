import nodemailer from "nodemailer";

interface EmailOptions {
  to: string;
  subject: string;
  text?: string;
  html?: string;
}

interface EmailResponse {
  success: boolean;
  messageId?: string;
  error?: string;
}

export const sendEmail = async (
  options: EmailOptions
): Promise<EmailResponse> => {
  try {
    // Validate required environment variables
    const email = process.env.EMAIL;
    const password = process.env.EMAIL_PASSWORD;

    if (!email || !password) {
      throw new Error(
        "EMAIL and EMAIL_PASSWORD environment variables are required"
      );
    }

    // Create transporter with Zoho Mail SMTP settings
    const transporter = nodemailer.createTransport({
      host: "smtp.zoho.com",
      port: 587,
      secure: false, // true for 465, false for other ports
      auth: {
        user: email,
        pass: password,
      },
      tls: {
        rejectUnauthorized: false,
      },
    });

    // Verify connection configuration
    await transporter.verify();

    // Send mail with defined transport object
    const info = await transporter.sendMail({
      from: `"Isaksweb Auth" <${email}>`, // sender address
      to: options.to, // list of receivers
      subject: options.subject, // Subject line
      text: options.text, // plain text body
      html: options.html, // html body
    });

    return {
      success: true,
      messageId: info.messageId,
    };
  } catch (error) {
    console.error("Error sending email:", error);
    return {
      success: false,
      error: error instanceof Error ? error.message : "Unknown error occurred",
    };
  }
};
