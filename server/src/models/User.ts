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
});

userSchema.pre("save", async function () {
  this.password = this.isModified("password")
    ? await bcrypt.hash(this.password, 10)
    : this.password;
});

export default mongoose.model("User", userSchema);
