import { useNavigate } from "react-router-dom";
import { Button } from "./Button";

/**
 * @interface BackButtonProps
 * @description 返回按钮组件的属性
 */
interface BackButtonProps {
  /** 返回的目标路径，如果不提供则使用浏览器历史记录返回上一页 */
  to?: string;
  /** 按钮显示的文本，默认为"返回" */
  label?: string;
  /** 按钮的变体样式 */
  variant?: "primary" | "secondary" | "danger" | "ghost";
  /** 自定义类名 */
  className?: string;
}

/**
 * @component BackButton
 * @description 通用的返回按钮组件
 * 支持返回到指定路径或浏览器历史记录的上一页
 */
export function BackButton({
  to,
  label = "返回",
  variant = "secondary",
  className = ""
}: BackButtonProps) {
  const navigate = useNavigate();

  const handleClick = () => {
    if (to) {
      navigate(to);
    } else {
      navigate(-1); // 返回上一页
    }
  };

  return (
    <Button
      variant={variant}
      onClick={handleClick}
      className={`back-button ${className}`}
    >
      ← {label}
    </Button>
  );
}

