const TransactionCard = ({ title, icon: Icon, color, children }) => {
  return (
    <div className="transaction-card">
      <div className="transaction-card-header">
        <div className="transaction-card-icon" style={{ color }}>
          <Icon size={28} />
        </div>
        <h3 className="transaction-card-title">{title}</h3>
      </div>
      <div className="transaction-card-divider" style={{ background: `linear-gradient(90deg, ${color}, transparent)` }}></div>
      <div className="transaction-card-body">
        {children}
      </div>
      <div className="corner-accent" style={{ borderColor: color }}></div>
    </div>
  );
};

export default TransactionCard;
