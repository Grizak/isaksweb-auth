import mongoose from "mongoose";
import bcrypt from "bcrypt";

const userSchema = new mongoose.Schema({
  email: {
    type: String,
    required: true,
    unique: true,
  },
  password: {
    type: String,
    required: true,
  },
  username: {
    type: String,
    required: true,
    unique: true,
  },
  age: {
    type: Number,
    required: true,
  },
});

userSchema.pre("save", async function () {
  this.password = this.isModified("password")
    ? await bcrypt.hash(this.password, 10)
    : this.password;
});

export default mongoose.model("User", userSchema);
