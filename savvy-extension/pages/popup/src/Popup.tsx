import '@src/Popup.css';
import { withSuspense } from '@extension/shared';

const Popup = () => {
  return (
    <div className="w-64 p-4">
      <div className="font-medium text-green-600">Logged in</div>
    </div>
  );
};

export default withSuspense(Popup, <div> Loading ... </div>);
