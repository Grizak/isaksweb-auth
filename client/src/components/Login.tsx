import { useState } from "react";
import type { JSX } from "react";
import { AlertCircle, CheckCircle2, Eye, EyeOff } from "lucide-react";
import { Link } from "react-router-dom";

// Types
interface FormData {
  [key: string]: string;
}

interface FormField {
  name: string;
  label: string;
  type: React.HTMLInputTypeAttribute;
  placeholder: string;
  showToggle?: boolean;
  matchField?: string;
  validation?: (value: string, formData: FormData) => string | null;
}

interface FormErrors {
  [key: string]: string;
}

type SubmitStatus = "success" | "error" | null;

// Define your form fields here
const FORM_FIELDS: FormField[] = [
  {
    name: "email",
    label: "Email",
    type: "email",
    placeholder: "your@email.com",
    validation: (value: string): string | null => {
      if (!value.trim()) return "Email is required";
      if (!/\S+@\S+\.\S+/.test(value)) return "Email is invalid";
      return null;
    },
  },
  {
    name: "password",
    label: "Password",
    type: "password",
    placeholder: "Enter your password",
    showToggle: true,
    validation: (value: string): string | null => {
      if (!value) return "Password is required";
      return null;
    },
  },
];

export default function LoginPage() {
  const initialFormData: FormData = FORM_FIELDS.reduce((acc, field) => {
    acc[field.name] = "";
    return acc;
  }, {} as FormData);

  const [formData, setFormData] = useState<FormData>(initialFormData);
  const [errors, setErrors] = useState<FormErrors>({});
  const [showPassword, setShowPassword] = useState<{ [key: string]: boolean }>(
    {}
  );
  const [isSubmitting, setIsSubmitting] = useState<boolean>(false);
  const [submitStatus, setSubmitStatus] = useState<SubmitStatus>(null);
  const [rememberMe, setRememberMe] = useState<boolean>(false);

  const validateForm = (): boolean => {
    const newErrors: FormErrors = {};

    FORM_FIELDS.forEach((field) => {
      if (field.validation) {
        const error = field.validation(formData[field.name], formData);
        if (error) {
          newErrors[field.name] = error;
        }
      }
    });

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const validateField = (fieldName: string, value: string): void => {
    const field = FORM_FIELDS.find((f) => f.name === fieldName);
    if (field?.validation) {
      const error = field.validation(value, formData);
      setErrors((prev) => ({
        ...prev,
        [fieldName]: error || "",
      }));
    }
  };

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>): void => {
    const { name, value } = e.target;
    setFormData((prev) => ({
      ...prev,
      [name]: value,
    }));

    // Clear error when user starts typing
    if (errors[name]) {
      setErrors((prev) => ({
        ...prev,
        [name]: "",
      }));
    }
  };

  const handleBlur = (e: React.FocusEvent<HTMLInputElement>): void => {
    const { name, value } = e.target;
    validateField(name, value);

    // Also validate matchField if this field is referenced by another
    const dependentField = FORM_FIELDS.find((f) => f.matchField === name);
    if (dependentField && formData[dependentField.name]) {
      validateField(dependentField.name, formData[dependentField.name]);
    }
  };

  const togglePasswordVisibility = (fieldName: string): void => {
    setShowPassword((prev) => ({
      ...prev,
      [fieldName]: !prev[fieldName],
    }));
  };

  const handleSubmit = async (
    e: React.MouseEvent<HTMLButtonElement>
  ): Promise<void> => {
    e.preventDefault();

    if (!validateForm()) {
      return;
    }

    setIsSubmitting(true);
    setSubmitStatus(null);

    try {
      // Filter out matchField entries from submission
      const submissionData = Object.keys(formData).reduce((acc, key) => {
        const field = FORM_FIELDS.find((f) => f.name === key);
        if (!field?.matchField) {
          acc[key] = formData[key];
        }
        return acc;
      }, {} as FormData);

      // Add rememberMe to submission
      const payload = {
        ...submissionData,
        rememberMe,
      };

      const response = await fetch("/api/login", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(payload),
      });

      if (response.ok) {
        setSubmitStatus("success");
        const data = await response.json();
        // Handle successful login (e.g., store token, redirect)
        console.log("Login successful:", data);
      } else {
        const data = await response.json();
        setSubmitStatus("error");
        setErrors({ submit: data.message || "Login failed" });
      }
    } catch {
      setSubmitStatus("error");
      setErrors({ submit: "Network error. Please try again." });
    } finally {
      setIsSubmitting(false);
    }
  };

  const renderField = (field: FormField): JSX.Element => {
    const fieldType =
      field.showToggle && showPassword[field.name] ? "text" : field.type;
    const hasError = errors[field.name];

    return (
      <div key={field.name}>
        <label
          htmlFor={field.name}
          className="block text-sm font-medium text-gray-700 mb-1"
        >
          {field.label}
        </label>
        <div className="relative">
          <input
            type={fieldType}
            id={field.name}
            name={field.name}
            value={formData[field.name]}
            onChange={handleChange}
            onBlur={handleBlur}
            className={`w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none transition ${
              field.showToggle ? "pr-10" : ""
            } ${hasError ? "border-red-500" : "border-gray-300"}`}
            placeholder={field.placeholder}
          />
          {field.showToggle && (
            <button
              type="button"
              onClick={() => togglePasswordVisibility(field.name)}
              className="absolute right-3 top-1/2 transform -translate-y-1/2 text-gray-500 hover:text-gray-700"
            >
              {showPassword[field.name] ? (
                <EyeOff size={20} />
              ) : (
                <Eye size={20} />
              )}
            </button>
          )}
        </div>
        {hasError && <p className="mt-1 text-sm text-red-600">{hasError}</p>}
      </div>
    );
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center p-4">
      <div className="bg-white rounded-lg shadow-xl w-full max-w-md p-8">
        <div className="text-center mb-8">
          <h1 className="text-3xl font-bold text-gray-900 mb-2">
            Welcome Back
          </h1>
          <p className="text-gray-600">Sign in to your account</p>
        </div>

        {submitStatus === "success" && (
          <div className="mb-6 p-4 bg-green-50 border border-green-200 rounded-lg flex items-start gap-3">
            <CheckCircle2
              className="text-green-600 flex-shrink-0 mt-0.5"
              size={20}
            />
            <div>
              <p className="text-green-800 font-medium">Login successful!</p>
              <p className="text-green-700 text-sm">Redirecting you now...</p>
            </div>
          </div>
        )}

        {errors.submit && (
          <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-lg flex items-start gap-3">
            <AlertCircle
              className="text-red-600 flex-shrink-0 mt-0.5"
              size={20}
            />
            <p className="text-red-800 text-sm">{errors.submit}</p>
          </div>
        )}

        <div className="space-y-5">
          {FORM_FIELDS.map((field) => renderField(field))}

          <div className="flex items-center justify-between">
            <label className="flex items-center">
              <input
                type="checkbox"
                checked={rememberMe}
                onChange={(e) => setRememberMe(e.target.checked)}
                className="w-4 h-4 text-blue-600 border-gray-300 rounded focus:ring-blue-500"
              />
              <span className="ml-2 text-sm text-gray-700">Remember me</span>
            </label>
            <Link
              to="/forgot-password"
              className="text-sm text-blue-600 hover:text-blue-700"
            >
              Forgot password?
            </Link>
          </div>

          <button
            type="button"
            onClick={handleSubmit}
            disabled={isSubmitting}
            className="w-full bg-blue-600 text-white py-2.5 px-4 rounded-lg font-medium hover:bg-blue-700 focus:ring-4 focus:ring-blue-300 transition disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {isSubmitting ? "Signing in..." : "Sign In"}
          </button>
        </div>

        <div className="mt-6 text-center">
          <p className="text-sm text-gray-600">
            Don't have an account?{" "}
            <Link
              to="/register"
              className="text-blue-600 hover:text-blue-700 font-medium"
            >
              Sign up
            </Link>
          </p>
          <Link
            to="/"
            className="text-blue-600 hover:text-blue-700 font-medium"
          >
            Or go back if you miss clicked
          </Link>
        </div>
      </div>
    </div>
  );
}
